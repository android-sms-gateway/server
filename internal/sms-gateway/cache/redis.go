package cache

import (
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) Cache {
	return &redisCache{
		client: client,
	}
}
