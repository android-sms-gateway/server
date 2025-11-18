package pubsub

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"pubsub",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("pubsub")
		}),
		fx.Provide(New),
		fx.Invoke(func(ps PubSub, logger *zap.Logger, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return nil
				},
				OnStop: func(_ context.Context) error {
					if err := ps.Close(); err != nil {
						logger.Error("pubsub close failed", zap.Error(err))
						return fmt.Errorf("failed to close pubsub: %w", err)
					}
					return nil
				},
			})
		}),
	)
}
