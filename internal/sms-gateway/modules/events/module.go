package events

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
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
							logger.Error("error running events service", zap.Error(err))
							if shErr := sh.Shutdown(fx.ExitCode(1)); shErr != nil {
								logger.Error("failed to shutdown", zap.Error(shErr))
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
}
