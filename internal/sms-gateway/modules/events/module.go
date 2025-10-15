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
	fx.Invoke(func(lc fx.Lifecycle, svc *Service, logger *zap.Logger, sh fx.Shutdowner) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					if err := svc.Run(ctx); err != nil {
						logger.Error("Error running events service", zap.Error(err))
						if err := sh.Shutdown(fx.ExitCode(1)); err != nil {
							logger.Error("Failed to shutdown", zap.Error(err))
						}
					}
				}()
				return nil
			},
			OnStop: func(_ context.Context) error {
				cancel()
				return nil
			},
		})
	}),
)
