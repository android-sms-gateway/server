package devices

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/db"
	"go.uber.org/zap"
)

type Service struct {
	config Config

	devices *Repository
	cache   *cache

	idGen db.IDGen

	logger *zap.Logger
}

func NewService(
	config Config,
	devices *Repository,
	idGen db.IDGen,
	logger *zap.Logger,
) *Service {
	return &Service{
		config: config,

		devices: devices,
		cache:   newCache(),

		idGen: idGen,

		logger: logger,
	}
}

func (s *Service) Insert(ctx context.Context, userID string, device DeviceInfo) (*Device, error) {
	input := DeviceInput{
		DeviceInfo: device,
		ID:         s.idGen(),
		UserID:     userID,
		AuthToken:  s.idGen(),
	}

	return s.devices.Insert(ctx, input)
}

// Select returns a list of devices for a specific user that match the provided filters.
func (s *Service) Select(ctx context.Context, userID string, filter ...SelectFilter) ([]Device, error) {
	filter = append(filter, WithUserID(userID))

	return s.devices.Select(ctx, filter...)
}

// Exists checks if there exists a device that matches the provided filters.
//
// If the device does not exist, it returns false and nil error. If there is an
// error during the query, it returns false and the error. Otherwise, it returns
// true and nil error.
func (s *Service) Exists(ctx context.Context, userID string, filter ...SelectFilter) (bool, error) {
	filter = append(filter, WithUserID(userID))

	return s.devices.Exists(ctx, filter...)
}

// Get returns a single device based on the provided filters for a specific user.
// It ensures that the filter includes the user's ID. If no device matches the
// criteria, it returns ErrNotFound. If more than one device matches, it returns
// ErrMoreThanOne.
func (s *Service) Get(ctx context.Context, userID string, filter ...SelectFilter) (*Device, error) {
	filter = append(filter, WithUserID(userID))

	return s.devices.Get(ctx, filter...)
}

func (s *Service) GetAny(ctx context.Context, userID string, deviceID string, duration time.Duration) (*Device, error) {
	filter := []SelectFilter{
		WithUserID(userID),
	}
	if deviceID != "" {
		filter = append(filter, WithID(deviceID))
	}
	if duration > 0 {
		filter = append(filter, ActiveWithin(duration))
	}

	devices, err := s.devices.Select(ctx, filter...)
	if err != nil {
		return nil, err
	}

	if len(devices) == 0 {
		return nil, ErrNotFound
	}

	if len(devices) == 1 {
		return &devices[0], nil
	}

	idx := rand.IntN(len(devices)) //nolint:gosec //not critical

	return &devices[idx], nil
}

// GetByToken returns a device by token.
//
// This method is used to retrieve a device by its auth token. If the device
// does not exist, it returns ErrNotFound.
func (s *Service) GetByToken(ctx context.Context, token string) (*Device, error) {
	device, err := s.cache.GetByToken(token)
	if err == nil {
		return &device, nil
	}

	devicePtr, err := s.devices.Get(ctx, WithToken(token))
	if err != nil {
		return nil, err
	}

	if setErr := s.cache.Set(*devicePtr); setErr != nil {
		s.logger.Error("failed to cache device", zap.String("device_id", devicePtr.ID), zap.Error(setErr))
	}

	return devicePtr, nil
}

func (s *Service) Update(ctx context.Context, id string, device DeviceUpdate) error {
	if err := s.devices.Update(ctx, id, device); err != nil {
		return err
	}

	if cacheErr := s.cache.DeleteByID(id); cacheErr != nil {
		s.logger.Error("failed to invalidate cache",
			zap.String("device_id", id),
			zap.Error(cacheErr),
		)
	}

	return nil
}

func (s *Service) SetLastSeen(ctx context.Context, batch map[string]time.Time) error {
	if err := s.devices.SetLastSeenBatch(ctx, batch); err != nil {
		s.logger.Error("failed to set last seen batch", zap.Error(err))
		return fmt.Errorf("failed to set last seen batch: %w", err)
	}
	return nil
}

// Remove removes devices for a specific user that match the provided filters.
// It ensures that the filter includes the user's ID.
func (s *Service) Remove(ctx context.Context, userID string, filter ...SelectFilter) error {
	filter = append(filter, WithUserID(userID))

	devices, err := s.devices.Select(ctx, filter...)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return nil
	}

	for _, device := range devices {
		if cacheErr := s.cache.DeleteByID(device.ID); cacheErr != nil {
			s.logger.Error("failed to invalidate cache",
				zap.String("device_id", device.ID),
				zap.Error(cacheErr),
			)
		}
	}

	if rmErr := s.devices.Remove(ctx, filter...); rmErr != nil {
		return rmErr
	}

	return nil
}
