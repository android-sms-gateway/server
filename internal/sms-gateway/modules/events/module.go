package events

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"events",
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("events")
	}),
	fx.Provide(newMetrics, fx.Private),
	fx.Provide(NewService),
	fx.Invoke(func(lc fx.Lifecycle, svc *Service) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go svc.Run(ctx)
				return nil
			},
			OnStop: func(_ context.Context) error {
				cancel()
				return nil
			},
		})
	}),
)
