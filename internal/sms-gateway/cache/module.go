package cache

import (
	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

type Factory cachefx.Factory

func Module() fx.Option {
	return fx.Module(
		"cache",
		logger.WithNamedLogger("cache"),
		fx.Provide(func(factory cachefx.Factory) Factory {
			return factory.WithName("sms-gateway")
		}),
	)
}
