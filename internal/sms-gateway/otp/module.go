package otp

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"otp",
		logger.WithNamedLogger("otp"),
		fx.Provide(
			func(factory cache.Factory) (cache.Cache, error) {
				return factory.New("otp")
			},
			NewStorage,
			fx.Private,
		),
		fx.Provide(NewService),
	)
}
