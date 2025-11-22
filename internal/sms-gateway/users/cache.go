package users

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/pkg/cache"
)

const loginCacheTTL = time.Hour

type loginCacheWrapper struct {
	ID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (w *loginCacheWrapper) Unmarshal(data []byte) error {
	return json.Unmarshal(data, w)
}

func (w *loginCacheWrapper) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

type loginCache struct {
	storage *cache.Typed[*loginCacheWrapper]
}

func newLoginCache(storage cache.Cache) *loginCache {
	return &loginCache{
		storage: cache.NewTyped[*loginCacheWrapper](storage),
	}
}

func (c *loginCache) makeKey(username, password string) string {
	hash := sha256.Sum256([]byte(username + "\x00" + password))
	return hex.EncodeToString(hash[:])
}

func (c *loginCache) Get(ctx context.Context, username, password string) (*User, error) {
	user, err := c.storage.Get(ctx, c.makeKey(username, password), cache.AndSetTTL(loginCacheTTL))
	if err != nil {
		return nil, fmt.Errorf("failed to get user from cache: %w", err)
	}

	return &User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (c *loginCache) Set(ctx context.Context, username, password string, user User) error {
	wrapper := &loginCacheWrapper{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if err := c.storage.Set(ctx, c.makeKey(username, password), wrapper, cache.WithTTL(loginCacheTTL)); err != nil {
		return fmt.Errorf("failed to cache user: %w", err)
	}

	return nil
}

func (c *loginCache) Delete(ctx context.Context, username, password string) error {
	err := c.storage.Delete(ctx, c.makeKey(username, password))
	if err == nil || errors.Is(err, cache.ErrKeyNotFound) || errors.Is(err, cache.ErrKeyExpired) {
		return nil
	}

	return fmt.Errorf("failed to delete user from cache: %w", err)
}
