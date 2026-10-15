package app

import (
	"github.com/google/gops/agent"
	"go-gin/internal/app/config"
)

// InitMonitor 初始化服务监控
func InitMonitor() error {
	if c := config.GetGlobalConfig().Monitor; c.Enable {
		err := agent.Listen(agent.Options{Addr: c.Addr, ConfigDir: c.ConfigDir, ShutdownCleanup: true})
		if err != nil {
			return err
		}
	}
	return nil
}
