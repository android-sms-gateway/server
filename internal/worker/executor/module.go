package executor

import (
	"github.com/go-core-fx/fxutil"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"executor",
		logger.WithNamedLogger("executor"),
		fx.Provide(newMetrics, fx.Private),
		fx.Provide(
			fx.Annotate(NewService, fx.ParamTags(`group:"worker:tasks"`)),
		),
		fx.Invoke(
			fxutil.RegisterRunnable[*Service](),
		),
	)
}

func AsWorkerTask(provider any) any {
	return fx.Annotate(provider, fx.ResultTags(`group:"worker:tasks"`))
}
