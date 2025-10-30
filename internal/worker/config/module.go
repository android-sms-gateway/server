package config

import (
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/worker/server"
	"github.com/android-sms-gateway/server/internal/worker/tasks/devices"
	"github.com/android-sms-gateway/server/internal/worker/tasks/messages"
	"github.com/capcom6/go-infra-fx/config"
	"github.com/capcom6/go-infra-fx/db"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"config",
		fx.Provide(func() (Config, error) {
			cfg := Default()

			if err := config.LoadConfig(&cfg); err != nil {
				return Config{}, fmt.Errorf("failed to load config: %w", err)
			}

			return cfg, nil
		}, fx.Private),
		fx.Provide(func(cfg Config) db.Config {
			return db.Config{
				Dialect:         db.DialectMySQL,
				DSN:             "",
				Host:            cfg.Database.Host,
				Port:            cfg.Database.Port,
				User:            cfg.Database.User,
				Password:        cfg.Database.Password,
				Database:        cfg.Database.Database,
				Timezone:        cfg.Database.Timezone,
				Debug:           cfg.Database.Debug,
				ConnMaxIdleTime: 0,
				ConnMaxLifetime: 0,
				MaxOpenConns:    cfg.Database.MaxOpenConns,
				MaxIdleConns:    cfg.Database.MaxIdleConns,
			}
		}),
		fx.Provide(func(cfg Config) messages.Config {
			return messages.Config{
				Hashing: messages.HashingConfig{
					Interval: time.Duration(cfg.Tasks.MessagesHashing.Interval),
				},
				Cleanup: messages.CleanupConfig{
					Interval: time.Duration(cfg.Tasks.MessagesCleanup.Interval),
					MaxAge:   time.Duration(cfg.Tasks.MessagesCleanup.MaxAge),
				},
			}
		}),
		fx.Provide(func(cfg Config) devices.Config {
			return devices.Config{
				Cleanup: devices.CleanupConfig{
					Interval: time.Duration(cfg.Tasks.DevicesCleanup.Interval),
					MaxAge:   time.Duration(cfg.Tasks.DevicesCleanup.MaxAge),
				},
			}
		}),
		fx.Provide(func(cfg Config) server.Config {
			return server.Config{
				Address: cfg.HTTP.Listen,
				Proxies: cfg.HTTP.Proxies,
			}
		}),
	)
}
