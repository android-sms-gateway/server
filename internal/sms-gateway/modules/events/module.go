package events

import (
	"github.com/go-core-fx/fxutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"events",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("events")
		}),
		fx.Provide(newMetrics, fx.Private),
		fx.Provide(NewService),
		fx.Invoke(
			fxutil.RegisterRunnable[*Service](),
		),
	)
}
