package messages

import (
	cacheFactory "github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/cleaner"
	"github.com/capcom6/go-infra-fx/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"messages",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("messages")
		}),
		fx.Provide(func(factory cacheFactory.Factory) (cacheFactory.Cache, error) {
			return factory.New("messages")
		}, fx.Private),

		fx.Provide(newMetrics, fx.Private),
		fx.Provide(newRepository, fx.Private),
		fx.Provide(newHashingTask, fx.Private),
		fx.Provide(newCache, fx.Private),

		fx.Provide(NewService),
		fx.Provide(
			cleaner.AsCleanable(
				func(svc *Service) cleaner.Cleanable {
					return svc
				},
			),
		),
	)
}

func init() {
	db.RegisterMigration(Migrate)
}
