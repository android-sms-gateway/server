package health

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"health",
		logger.WithNamedLogger("health"),
		fx.Provide(
			AsHealthProvider(newHealth),
			fx.Private,
		),
		fx.Provide(
			fx.Annotate(NewService, fx.ParamTags(`group:"health-providers"`)),
		),
	)
}

func AsHealthProvider(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Provider)),
		fx.ResultTags(`group:"health-providers"`),
	)
}
