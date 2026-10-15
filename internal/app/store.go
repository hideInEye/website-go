package app

import (
	"errors"
	"os"
	"path/filepath"

	"go-gin/internal/app/config"
	igorm "go-gin/internal/app/model/impl/gorm"
	"go-gin/pkg/gormplus"

	"github.com/jinzhu/gorm"
	"go.uber.org/dig"
)

// InitStore 初始化存储
func InitStore(container *dig.Container) (func(), error) {
	var storeCall func()
	cfg := config.GetGlobalConfig()

	switch cfg.Store {
	case "gorm":
		db, err := initGorm()
		if err != nil {
			return nil, err
		}

		storeCall = func() {
			db.Close()
		}

		igorm.SetTablePrefix(cfg.Gorm.TablePrefix)
		if cfg.Gorm.AutoMigrate {
			err = igorm.AutoMigrate(db)
			if err != nil {
				return nil, err
			}
		}

		// 注入DB
		err = container.Provide(func() *gormplus.DB {
			return db
		})
		if err != nil {
			db.Close()
			return nil, err
		}

		// 输入gorm.DB
		err = container.Provide(func() *gorm.DB {
			return db.GetDB()
		})
		if err != nil {
			db.Close()
			return nil, err
		}

		if err := igorm.Inject(container); err != nil {
			db.Close()
			return nil, err
		}
	default:
		return nil, errors.New("unknown store")
	}

	return storeCall, nil
}

// initGorm 实例化gorm存储
func initGorm() (*gormplus.DB, error) {
	cfg := config.GetGlobalConfig()

	var dsn string
	switch cfg.Gorm.DBType {
	case "mysql":
		dsn = cfg.MySQL.DSN()
	case "sqlite3":
		dsn = cfg.Sqlite3.DSN()
		os.MkdirAll(filepath.Dir(dsn), 0777)
	case "postgres":
		dsn = cfg.Postgres.DSN()
	default:
		return nil, errors.New("unknown db")
	}

	return gormplus.New(&gormplus.Config{
		Debug:        cfg.Gorm.Debug,
		DBType:       cfg.Gorm.DBType,
		DSN:          dsn,
		MaxIdleConns: cfg.Gorm.MaxIdleConns,
		MaxLifetime:  cfg.Gorm.MaxLifetime,
		MaxOpenConns: cfg.Gorm.MaxOpenConns,
	})
}
