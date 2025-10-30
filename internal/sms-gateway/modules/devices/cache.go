package devices

import (
	"errors"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	cacheImpl "github.com/capcom6/go-helpers/cache"
)

type cache struct {
	byID    *cacheImpl.Cache[*models.Device]
	byToken *cacheImpl.Cache[*models.Device]
}

func newCache() *cache {
	const ttl = 10 * time.Minute

	return &cache{
		byID:    cacheImpl.New[*models.Device](cacheImpl.Config{TTL: ttl}),
		byToken: cacheImpl.New[*models.Device](cacheImpl.Config{TTL: ttl}),
	}
}

func (c *cache) Set(device models.Device) error {
	err := errors.Join(c.byID.Set(device.ID, &device), c.byToken.Set(device.AuthToken, &device))
	if err != nil {
		return fmt.Errorf("failed to cache device: %w", err)
	}

	return nil
}

func (c *cache) GetByID(id string) (models.Device, error) {
	device, err := c.byID.Get(id)
	if err != nil {
		return models.Device{}, fmt.Errorf("failed to get device by ID: %w", err)
	}

	return *device, nil
}

func (c *cache) GetByToken(token string) (models.Device, error) {
	device, err := c.byToken.Get(token)
	if err != nil {
		return models.Device{}, fmt.Errorf("failed to get device by token: %w", err)
	}

	return *device, nil
}

func (c *cache) DeleteByID(id string) error {
	device, err := c.byID.Get(id)
	if err != nil {
		if errors.Is(err, cacheImpl.ErrKeyNotFound) {
			return nil
		}

		return fmt.Errorf("failed to get device by ID: %w", err)
	}

	err = errors.Join(c.byID.Delete(device.ID), c.byToken.Delete(device.AuthToken))
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	return nil
}
