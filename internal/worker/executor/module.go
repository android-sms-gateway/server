package executor

import (
	"context"

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
		fx.Invoke(func(svc *Service, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return svc.Start()
				},
				OnStop: func(_ context.Context) error {
					return svc.Stop()
				},
			})
		}),
	)
}

func AsWorkerTask(provider any) any {
	return fx.Annotate(provider, fx.ResultTags(`group:"worker:tasks"`))
}
