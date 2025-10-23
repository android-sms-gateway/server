package health

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"health",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("health")
		}),
		fx.Provide(
			AsHealthProvider(NewHealth),
			fx.Private,
		),
		fx.Provide(
			NewService,
		),
	)
}

func AsHealthProvider(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(HealthProvider)),
		fx.ResultTags(`group:"health-providers"`),
	)
}
