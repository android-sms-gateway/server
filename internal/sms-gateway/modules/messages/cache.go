package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cacheImpl "github.com/go-core-fx/cachefx/cache"
)

const (
	cacheTimeout = 100 * time.Millisecond
)

type cache struct {
	ttl time.Duration

	storage cacheImpl.Cache
}

func newCache(config Config, storage cacheImpl.Cache) *cache {
	return &cache{
		ttl: config.CacheTTL,

		storage: storage,
	}
}

func (c *cache) Set(ctx context.Context, userID, id string, message *MessageStateOut) error {
	var (
		err  error
		data []byte
	)

	if message != nil {
		data, err = json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	if setErr := c.storage.Set(ctx, userID+":"+id, data, cacheImpl.WithTTL(c.ttl)); setErr != nil {
		return fmt.Errorf("failed to set message in cache: %w", setErr)
	}

	return nil
}

func (c *cache) Get(ctx context.Context, userID, id string) (*MessageStateOut, error) {
	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	data, err := c.storage.Get(ctx, userID+":"+id, cacheImpl.AndSetTTL(c.ttl))
	if err != nil {
		return nil, fmt.Errorf("failed to get message from cache: %w", err)
	}

	if len(data) == 0 {
		return nil, nil //nolint:nilnil //empty cached value is used for caching "Not Found"
	}

	message := new(MessageStateOut)
	if jsonErr := json.Unmarshal(data, message); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", jsonErr)
	}

	return message, nil
}
