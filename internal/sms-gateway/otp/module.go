package otp

import (
	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/cachefx/cache"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"otp",
		logger.WithNamedLogger("otp"),
		fx.Provide(
			func(factory cachefx.Factory) (cache.Cache, error) {
				return factory.New("otp")
			},
			NewStorage,
			fx.Private,
		),
		fx.Provide(NewService),
	)
}
