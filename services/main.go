package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go-gin/internal/app"
	"go-gin/internal/app/config"
	store "go-gin/internal/app/model/impl/gorm"
	"go-gin/pkg/logger"
	"go-gin/until"
)

var VERSION = "4.0.0"

var (
	configFile string
	modelFile  string
	wwwDir     string
	swaggerDir string
)

func init() {
	flag.StringVar(&configFile, "c", "config/config.toml", "TOML 配置文件")
	flag.StringVar(&modelFile, "m", "", "Casbin的访问控制模型(.conf)")
	flag.StringVar(&wwwDir, "www", "", "静态站点目录")
	flag.StringVar(&swaggerDir, "swagger", "", "swagger目录")
}

func main() {
	command := "serve"
	args := os.Args[1:]
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		command = args[0]
		args = args[1:]
	}
	if err := flag.CommandLine.Parse(args); err != nil {
		failCommand(err)
	}
	if configFile == "" {
		failCommand(fmt.Errorf("请使用 -c 指定 TOML 配置文件"))
	}
	switch command {
	case "migrate":
		cfg, err := config.ParseConfig(configFile)
		if err != nil {
			failCommand(err)
		}
		if err := cfg.ValidateDatabase(); err != nil {
			failCommand(err)
		}
		if err := store.Migrate(*cfg); err != nil {
			failCommand(err)
		}
		return
	case "healthcheck":
		cfg, err := config.ParseConfig(configFile)
		if err != nil {
			failCommand(err)
		}
		if err := healthcheck(cfg); err != nil {
			failCommand(err)
		}
		return
	case "serve":
	default:
		failCommand(fmt.Errorf("unknown command: %s", command))
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(signals)

	ctx := logger.NewTraceIDContext(context.Background(), until.MustUUID())
	span := logger.StartSpanWithCall(ctx)
	cleanup := app.Init(ctx,
		app.SetConfigFile(configFile),
		app.SetModelFile(modelFile),
		app.SetSwaggerDir(swaggerDir),
		app.SetWWWDir(wwwDir),
		app.SetVersion(VERSION),
	)
	signal := <-signals
	span().Printf("获取到退出信号[%s]", signal.String())
	if cleanup != nil {
		cleanup()
	}
	span().Printf("服务退出")
}

func healthcheck(cfg *config.Config) error {
	host := cfg.HTTP.Host
	switch host {
	case "", "0.0.0.0":
		host = "127.0.0.1"
	case "::":
		host = "::1"
	}
	address := net.JoinHostPort(host, strconv.Itoa(cfg.HTTP.Port))
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Get("http://" + address + "/api/v1/health/live")
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck returned HTTP %d", response.StatusCode)
	}
	return nil
}

func failCommand(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
