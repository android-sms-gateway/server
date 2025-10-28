package pubsub

import (
	"context"

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
				OnStop: func(_ context.Context) error {
					if err := ps.Close(); err != nil {
						logger.Error("pubsub close failed", zap.Error(err))
						return err
					}
					return nil
				},
			})
		}),
	)
}
