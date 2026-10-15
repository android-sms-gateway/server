package cache

import (
	"sync"

	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/cachefx/cache"
	"github.com/go-core-fx/fxutil"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"cache",
		logger.WithNamedLogger("cache"),
		fx.Provide(func(factory cachefx.Factory) Factory {
			return &trackingFactory{
				Factory: factory.WithName("sms-gateway"),

				cachesMu: sync.Mutex{},
				caches:   []cache.Cache{},
			}
		}),
		fx.Provide(NewSweeper, fx.Private),
		fx.Invoke(
			fxutil.RegisterRunnable[*Sweeper](),
		),
	)
}
