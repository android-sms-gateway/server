package config

import (
	"strings"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/auth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/sse"
	"github.com/android-sms-gateway/server/internal/sms-gateway/pubsub"
	"github.com/capcom6/go-infra-fx/config"
	"github.com/capcom6/go-infra-fx/db"
	"github.com/capcom6/go-infra-fx/http"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

//nolint:funlen // long function
func Module() fx.Option {
	return fx.Module(
		"appconfig",
		fx.Provide(
			func(log *zap.Logger) Config {
				defaultConfig := Default()

				if err := config.LoadConfig(&defaultConfig); err != nil {
					log.Error("Error loading config", zap.Error(err))
				}

				return defaultConfig
			},
			fx.Private,
		),
		fx.Provide(func(cfg Config) http.Config {
			const writeTimeout = 30 * time.Minute

			return http.Config{
				Listen:  cfg.HTTP.Listen,
				Proxies: cfg.HTTP.Proxies,

				WriteTimeout: writeTimeout, // SSE requires longer timeout
			}
		}),
		fx.Provide(func(cfg Config) db.Config {
			return db.Config{
				Dialect:  db.DialectMySQL,
				Host:     cfg.Database.Host,
				Port:     cfg.Database.Port,
				User:     cfg.Database.User,
				Password: cfg.Database.Password,
				Database: cfg.Database.Database,
				Timezone: cfg.Database.Timezone,
				Debug:    cfg.Database.Debug,

				MaxOpenConns: cfg.Database.MaxOpenConns,
				MaxIdleConns: cfg.Database.MaxIdleConns,

				DSN:             "",
				ConnMaxIdleTime: 0,
				ConnMaxLifetime: 0,
			}
		}),
		fx.Provide(func(cfg Config) push.Config {
			mode := push.ModeFCM
			if cfg.Gateway.Mode == GatewayModePrivate {
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
		fx.Provide(func(cfg Config) auth.Config {
			return auth.Config{
				Mode:         auth.Mode(cfg.Gateway.Mode),
				PrivateToken: cfg.Gateway.PrivateToken,
			}
		}),
		fx.Provide(func(cfg Config) handlers.Config {
			// Default and normalize API path/host
			if cfg.HTTP.API.Host == "" {
				cfg.HTTP.API.Path = "/api"
			}
			// Ensure leading slash and trim trailing slash (except root)
			if !strings.HasPrefix(cfg.HTTP.API.Path, "/") {
				cfg.HTTP.API.Path = "/" + cfg.HTTP.API.Path
			}
			if cfg.HTTP.API.Path != "/" && strings.HasSuffix(cfg.HTTP.API.Path, "/") {
				cfg.HTTP.API.Path = strings.TrimRight(cfg.HTTP.API.Path, "/")
			}
			// Guard against misconfigured scheme in host (accept "host[:port]" only)
			cfg.HTTP.API.Host = strings.TrimPrefix(strings.TrimPrefix(cfg.HTTP.API.Host, "https://"), "http://")

			return handlers.Config{
				PublicHost:      cfg.HTTP.API.Host,
				PublicPath:      cfg.HTTP.API.Path,
				UpstreamEnabled: cfg.Gateway.Mode == GatewayModePublic,
				OpenAPIEnabled:  cfg.HTTP.OpenAPI.Enabled,
			}
		}),
		fx.Provide(func(cfg Config) messages.Config {
			return messages.Config{
				CacheTTL: time.Duration(cfg.Messages.CacheTTLSeconds) * time.Second,
				HashingInterval: time.Duration(
					max(cfg.Tasks.Hashing.IntervalSeconds, cfg.Messages.HashingIntervalSeconds),
				) * time.Second,
			}
		}),
		fx.Provide(func(_ Config) devices.Config {
			return devices.Config{
				UnusedLifetime: 365 * 24 * time.Hour, //TODO: make it configurable
			}
		}),
		fx.Provide(func(cfg Config) sse.Config {
			return sse.NewConfig(
				sse.WithKeepAlivePeriod(time.Duration(cfg.SSE.KeepAlivePeriodSeconds) * time.Second),
			)
		}),
		fx.Provide(func(cfg Config) cache.Config {
			return cache.Config{
				URL: cfg.Cache.URL,
			}
		}),
		fx.Provide(func(cfg Config) pubsub.Config {
			return pubsub.Config{
				URL:        cfg.PubSub.URL,
				BufferSize: cfg.PubSub.BufferSize,
			}
		}),
	)
}
