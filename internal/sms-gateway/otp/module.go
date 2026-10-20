package otp

import (
	appCache "github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/go-core-fx/cachefx/cache"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"otp",
		logger.WithNamedLogger("otp"),
		fx.Provide(
			// The cache is created through the tracked factory, so the sweeper
			// reclaims the expired one-time codes.
			func(factory appCache.Factory) (cache.Cache, error) {
				return factory.New("otp")
			},
			NewStorage,
			fx.Private,
		),
		fx.Provide(NewService),
	)
}
