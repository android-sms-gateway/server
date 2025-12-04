package online

import (
	"context"

	appCache "github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/go-core-fx/cachefx/cache"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"online",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("online")
		}),
		fx.Provide(func(factory appCache.Factory) (cache.Cache, error) {
			return factory.New("online")
		}, fx.Private),
		fx.Provide(newMetrics),
		fx.Provide(New),
		fx.Invoke(func(lc fx.Lifecycle, svc Service) {
			ctx, cancel := context.WithCancel(context.Background())
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					go svc.Run(ctx)
					return nil
				},
				OnStop: func(_ context.Context) error {
					cancel()
					return nil
				},
			})
		}),
	)
}
