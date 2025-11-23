package thirdparty

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"thirdparty",
		logger.WithNamedLogger("3rdparty"),
		fx.Provide(
			NewAuthHandler,
		),
	)
}
