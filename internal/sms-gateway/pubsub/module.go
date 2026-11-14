package pubsub

import (
	"github.com/go-core-fx/logger"
	"github.com/go-core-fx/pubsubfx"
	"go.uber.org/fx"
)

// topicPrefix namespaces every topic this application publishes, so instances
// sharing one Redis server do not see each other's messages.
const topicPrefix = "sms-gateway:"

// Module maps the application config onto pubsubfx.Config. The PubSub itself
// is provided by pubsubfx.Module(), which also closes it on app stop.
func Module() fx.Option {
	return fx.Module(
		"pubsub",
		logger.WithNamedLogger("pubsub"),
		fx.Provide(func(cfg Config) pubsubfx.Config {
			return pubsubfx.Config{
				URL:        cfg.URL,
				BufferSize: cfg.BufferSize,
				Prefix:     topicPrefix,
			}
		}),
	)
}
