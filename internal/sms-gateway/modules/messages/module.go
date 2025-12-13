package messages

import (
	cacheFactory "github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/capcom6/go-infra-fx/db"
	cacheImpl "github.com/go-core-fx/cachefx/cache"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"messages",
		logger.WithNamedLogger("messages"),
		fx.Provide(func(factory cacheFactory.Factory) (cacheImpl.Cache, error) {
			return factory.New("messages")
		}, fx.Private),

		fx.Provide(newMetrics, fx.Private),
		fx.Provide(NewRepository, fx.Private),
		fx.Provide(newHashingWorker, fx.Private),
		fx.Provide(newCache, fx.Private),

		fx.Provide(NewService),
	)
}

//nolint:gochecknoinits //backward compatibility
func init() {
	db.RegisterMigration(Migrate)
}
