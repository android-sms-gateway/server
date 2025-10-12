package pubsub

import (
	"context"

	"github.com/android-sms-gateway/server/pkg/pubsub"
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
		fx.Invoke(func(pubsub pubsub.PubSub, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStop: func(_ context.Context) error {
					return pubsub.Close()
				},
			})
		}),
	)
}
