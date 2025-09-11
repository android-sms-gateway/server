package cache

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"cache",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("cache")
		}),
		fx.Provide(NewFactory),
	)
}
