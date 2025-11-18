package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/capcom6/go-helpers/cache"
)

type usersCache struct {
	cache *cache.Cache[models.User]
}

func newUsersCache() *usersCache {
	return &usersCache{
		cache: cache.New[models.User](cache.Config{TTL: 1 * time.Hour}),
	}
}

func (c *usersCache) makeKey(username, password string) string {
	hash := sha256.Sum256([]byte(username + "\x00" + password))
	return hex.EncodeToString(hash[:])
}

func (c *usersCache) Get(username, password string) (models.User, error) {
	user, err := c.cache.Get(c.makeKey(username, password))
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get user from cache: %w", err)
	}

	return user, nil
}

func (c *usersCache) Set(username, password string, user models.User) error {
	if err := c.cache.Set(c.makeKey(username, password), user); err != nil {
		return fmt.Errorf("failed to cache user: %w", err)
	}

	return nil
}

func (c *usersCache) Delete(username, password string) error {
	if err := c.cache.Delete(c.makeKey(username, password)); err != nil {
		return fmt.Errorf("failed to delete user from cache: %w", err)
	}

	return nil
}

func (c *usersCache) Cleanup() {
	c.cache.Cleanup()
}
