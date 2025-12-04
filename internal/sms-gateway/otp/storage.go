package otp

import (
	"context"
	"fmt"

	"github.com/go-core-fx/cachefx/cache"
)

// Storage wraps the cache interface for string values.
type Storage struct {
	cache cache.Cache
}

// NewStorage creates a new Storage with the underlying cache.
func NewStorage(c cache.Cache) *Storage {
	return &Storage{cache: c}
}

// SetOrFail is like Set, but returns an error if the key already exists.
func (s *Storage) SetOrFail(ctx context.Context, key string, value string, opts ...cache.Option) error {
	if err := s.cache.SetOrFail(ctx, key, []byte(value), opts...); err != nil {
		return fmt.Errorf("failed to set cache item: %w", err)
	}

	return nil
}

// GetAndDelete retrieves and deletes a string value.
func (s *Storage) GetAndDelete(ctx context.Context, key string) (string, error) {
	data, err := s.cache.GetAndDelete(ctx, key)
	return string(data), err
}

// Cleanup removes expired items.
func (s *Storage) Cleanup(ctx context.Context) error {
	if err := s.cache.Cleanup(ctx); err != nil {
		return fmt.Errorf("failed to cleanup cache: %w", err)
	}
	return nil
}
