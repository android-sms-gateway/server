package messages

import (
	cacheFactory "github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/capcom6/go-infra-fx/db"
	"github.com/go-core-fx/cachefx/cache"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"messages",
		logger.WithNamedLogger("messages"),
		fx.Provide(
			func(factory cacheFactory.Factory) (cache.Cache, error) {
				return factory.New("messages")
			},
			func(config Config) QueueConfig {
				return config.Queue
			},
			fx.Private,
		),

		fx.Provide(newMetrics, fx.Private),
		fx.Provide(newHashingWorker, fx.Private),
		fx.Provide(newCache, fx.Private),

		fx.Provide(
			NewRepository,
			NewLimiter,
			fx.Private,
		),
		fx.Provide(NewService),
	)
}

//nolint:gochecknoinits //backward compatibility
func init() {
	db.RegisterMigration(Migrate)
}
