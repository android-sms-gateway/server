package cache

import (
	"fmt"
	"net/url"

	"github.com/android-sms-gateway/server/pkg/cache"
)

const (
	keyPrefix = "sms-gateway:"
)

type Cache = cache.Cache

type Factory interface {
	New(name string) (Cache, error)
}

type factory struct {
	new func(name string) (Cache, error)
}

func NewFactory(config Config) (Factory, error) {
	if config.URL == "" {
		config.URL = "memory://"
	}

	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("can't parse url: %w", err)
	}

	switch u.Scheme {
	case "memory":
		return &factory{
			new: func(name string) (Cache, error) {
				return cache.NewMemory(0), nil
			},
		}, nil
	case "redis":
		return &factory{
			new: func(name string) (Cache, error) {
				return cache.NewRedis(cache.RedisConfig{
					Client: nil,
					URL:    config.URL,
					Prefix: keyPrefix + name,
					TTL:    0,
				})
			},
		}, nil
	default:
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
}

// New implements Factory.
func (f *factory) New(name string) (Cache, error) {
	return f.new(name)
}
