package health

import (
	"context"
	"time"

	"github.com/android-sms-gateway/server/internal/config"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Run() {
	fx.New(
		fx.StartTimeout(time.Second),
		logger.Module(),
		logger.WithFxDefaultLogger(),
		config.Module(),
		module(),
	).Run()
}

func module() fx.Option {
	return fx.Module(
		"health",
		fx.Provide(NewChecker),
		fx.Invoke(func(lc fx.Lifecycle, checker *Checker) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return checker.Execute(ctx)
				},
				OnStop: func(_ context.Context) error {
					return nil
				},
			})
		}),
	)
}
