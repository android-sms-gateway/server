package cache

import (
	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

// Factory is the application cache factory: cachefx's factory scoped to the
// "sms-gateway" prefix.
//
// Caches created through it are swept by the cachefx sweeper, because
// WithName shares the sweeper's tracker with the root factory.
type Factory interface {
	cachefx.Factory
}

func Module() fx.Option {
	return fx.Module(
		"cache",
		logger.WithNamedLogger("cache"),
		fx.Provide(func(factory cachefx.Factory) Factory {
			return factory.WithName("sms-gateway")
		}),
	)
}
