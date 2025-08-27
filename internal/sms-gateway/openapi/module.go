package openapi

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"openapi",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("openapi")
		}),
		fx.Provide(New),
	)
}
