package auth

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"auth",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("auth")
		}),
		fx.Provide(New),
	)
}
