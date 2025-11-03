package pubsub

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/android-sms-gateway/server/pkg/pubsub"
)

const (
	topicPrefix = "sms-gateway:"
)

type PubSub = pubsub.PubSub

var ErrInvalidScheme = errors.New("invalid scheme")

func New(config Config) (PubSub, error) {
	if config.URL == "" {
		config.URL = "memory://"
	}

	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	opts := []pubsub.Option{}
	opts = append(opts, pubsub.WithBufferSize(config.BufferSize))

	var pubSub PubSub
	switch u.Scheme {
	case "memory":
		pubSub, err = pubsub.NewMemory(opts...), nil
	case "redis":
		pubSub, err = pubsub.NewRedis(pubsub.RedisConfig{
			Client: nil,
			URL:    config.URL,
			Prefix: topicPrefix,
		}, opts...)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidScheme, u.Scheme)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	return pubSub, nil
}
