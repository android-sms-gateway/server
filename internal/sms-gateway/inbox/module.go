package inbox

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"inbox",
		logger.WithNamedLogger("inbox"),
		fx.Provide(
			New,
		),
	)
}
