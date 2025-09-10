package cache

import (
	"fmt"
	"net/url"

	"github.com/android-sms-gateway/core/redis"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"cache",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("cache")
		}),
		fx.Provide(func(cfg Config) (Cache, error) {
			u, err := url.Parse(cfg.URL)
			if err != nil {
				return nil, fmt.Errorf("can't parse url: %w", err)
			}
			switch u.Scheme {
			case "memory":
				return NewMemory(), nil
			case "redis":
				client, err := redis.New(redis.Config{URL: cfg.URL})
				if err != nil {
					return nil, fmt.Errorf("can't create redis client: %w", err)
				}
				return NewRedis(client), nil
			default:
				return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
			}
		}),
	)
}
