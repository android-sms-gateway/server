package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cacheImpl "github.com/android-sms-gateway/server/pkg/cache"
)

type cache struct {
	storage cacheImpl.Cache
}

func newCache(storage cacheImpl.Cache) *cache {
	return &cache{
		storage: storage,
	}
}

func (c *cache) Set(ctx context.Context, userID, ID string, message *MessageStateOut) error {
	var (
		err  error
		data []byte
	)

	if message != nil {
		data, err = json.Marshal(message)
		if err != nil {
			return fmt.Errorf("can't marshal message: %w", err)
		}
	}

	return c.storage.Set(ctx, userID+":"+ID, data, cacheImpl.WithTTL(5*time.Minute))
}

func (c *cache) Get(ctx context.Context, userID, ID string) (*MessageStateOut, error) {
	data, err := c.storage.Get(ctx, userID+":"+ID)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	message := new(MessageStateOut)
	if err := json.Unmarshal(data, message); err != nil {
		return nil, fmt.Errorf("can't unmarshal message: %w", err)
	}

	return message, nil
}
