package devices

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"devices",
		logger.WithNamedLogger("devices"),
		fx.Provide(
			NewRepository,
			fx.Private,
		),
		fx.Provide(NewService),
	)
}
