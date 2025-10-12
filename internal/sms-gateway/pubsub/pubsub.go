package pubsub

import (
	"fmt"
	"net/url"

	"github.com/android-sms-gateway/server/pkg/pubsub"
)

const (
	topicPrefix = "sms-gateway:"
)

func New(config Config) (pubsub.PubSub, error) {
	if config.URL == "" {
		config.URL = "memory://"
	}

	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("can't parse url: %w", err)
	}

	opts := []pubsub.Option{}
	opts = append(opts, pubsub.WithBufferSize(config.BufferSize))

	switch u.Scheme {
	case "memory":
		return pubsub.NewMemory(opts...), nil
	case "redis":
		return pubsub.NewRedis(pubsub.RedisConfig{
			Client: nil,
			URL:    config.URL,
			Prefix: topicPrefix,
		}, opts...)
	default:
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
}
