package config

import (
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/auth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
	"github.com/capcom6/go-infra-fx/config"
	"github.com/capcom6/go-infra-fx/db"
	"github.com/capcom6/go-infra-fx/http"
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
			Listen:  cfg.HTTP.Listen,
			Proxies: cfg.HTTP.Proxies,
		}
	}),
	fx.Provide(func(cfg Config) db.Config {
		return db.Config{
			Dialect:  db.Dialect(cfg.Database.Dialect),
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Database: cfg.Database.Database,
			Timezone: cfg.Database.Timezone,
			Debug:    cfg.Database.Debug,

			MaxOpenConns: cfg.Database.MaxOpenConns,
			MaxIdleConns: cfg.Database.MaxIdleConns,
		}
	}),
	fx.Provide(func(cfg Config) push.Config {
		mode := push.ModeFCM
		if cfg.Gateway.Mode == "private" {
			mode = push.ModeUpstream
		}

		return push.Config{
			Mode: mode,
			ClientOptions: map[string]string{
				"credentials": cfg.FCM.CredentialsJSON,
			},
			Debounce: time.Duration(cfg.FCM.DebounceSeconds) * time.Second,
			Timeout:  time.Duration(cfg.FCM.TimeoutSeconds) * time.Second,
		}
	}),
	fx.Provide(func(cfg Config) messages.HashingTaskConfig {
		return messages.HashingTaskConfig{
			Interval: time.Duration(cfg.Tasks.Hashing.IntervalSeconds) * time.Second,
		}
	}),
	fx.Provide(func(cfg Config) auth.Config {
		return auth.Config{
			Mode:         auth.Mode(cfg.Gateway.Mode),
			PrivateToken: cfg.Gateway.PrivateToken,
		}
	}),
	fx.Provide(func(cfg Config) handlers.Config {
		return handlers.Config{
			GatewayMode: handlers.GatewayMode(cfg.Gateway.Mode),
		}
	}),
	fx.Provide(func(cfg Config) messages.Config {
		return messages.Config{
			ProcessedLifetime: 30 * 24 * time.Hour, //TODO: make it configurable
		}
	}),
	fx.Provide(func(cfg Config) devices.Config {
		return devices.Config{
			UnusedLifetime: 365 * 24 * time.Hour, //TODO: make it configurable
		}
	}),
)
