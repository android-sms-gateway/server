package cache

import (
	"fmt"
	"net/url"

	"github.com/android-sms-gateway/core/redis"
	"github.com/android-sms-gateway/server/pkg/cache"
)

type Factory interface {
	New(name string) (cache.Cache, error)
}

type factory struct {
	new func(name string) (cache.Cache, error)
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
			new: func(name string) (cache.Cache, error) {
				return cache.NewMemory(0), nil
			},
		}, nil
	case "redis":
		client, err := redis.New(redis.Config{URL: config.URL})
		if err != nil {
			return nil, fmt.Errorf("can't create redis client: %w", err)
		}
		return &factory{
			new: func(name string) (cache.Cache, error) {
				return cache.NewRedis(client, name, 0), nil
			},
		}, nil
	default:
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
}

// New implements Factory.
func (f *factory) New(name string) (cache.Cache, error) {
	return f.new(name)
}
