package config

import (
	"time"

	"github.com/capcom6/go-infra-fx/config"
	"github.com/capcom6/go-infra-fx/db"
	"github.com/capcom6/go-infra-fx/http"
	"github.com/capcom6/sms-gateway/internal/sms-gateway/services"
	"github.com/capcom6/sms-gateway/internal/sms-gateway/tasks"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"appconfig",
	fx.Provide(
		func(log *zap.Logger) Config {
			if err := config.LoadConfig(&defaultConfig); err != nil {
				log.Error("Error loading config", zap.Error(err))
			}

			return defaultConfig
		},
	),
	fx.Provide(func(cfg Config) http.Config {
		return http.Config{
			Listen: cfg.HTTP.Listen,
		}
	}),
	fx.Provide(func(cfg Config) db.Config {
		return db.Config{
			Dialect:  cfg.Database.Dialect,
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Database: cfg.Database.Database,
			Timezone: cfg.Database.Timezone,
		}
	}),
	fx.Provide(func(cfg Config) services.PushServiceConfig {
		return services.PushServiceConfig{
			CredentialsJSON: cfg.FCM.CredentialsJSON,
			Debounce:        time.Duration(cfg.FCM.DebounceSeconds) * time.Second,
			Timeout:         time.Duration(cfg.FCM.TimeoutSeconds) * time.Second,
		}
	}),
	fx.Provide(func(cfg Config) tasks.HashingTaskConfig {
		return tasks.HashingTaskConfig{
			Interval: time.Duration(cfg.Tasks.Hashing.IntervalSeconds) * time.Second,
		}
	}),
)
