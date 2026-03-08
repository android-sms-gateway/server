package online

import (
	appCache "github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/go-core-fx/cachefx/cache"
	"github.com/go-core-fx/fxutil"
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
		fx.Invoke(
			fxutil.RegisterRunnable[Service](),
		),
	)
}
