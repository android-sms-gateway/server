package devices

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
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

func (s *Service) Insert(userID string, device *models.Device) error {
	device.ID = s.idGen()
	device.AuthToken = s.idGen()
	device.UserID = userID

	return s.devices.Insert(device)
}

// Select returns a list of devices for a specific user that match the provided filters.
func (s *Service) Select(userID string, filter ...SelectFilter) ([]models.Device, error) {
	filter = append(filter, WithUserID(userID))

	return s.devices.Select(filter...)
}

// Exists checks if there exists a device that matches the provided filters.
//
// If the device does not exist, it returns false and nil error. If there is an
// error during the query, it returns false and the error. Otherwise, it returns
// true and nil error.
func (s *Service) Exists(userID string, filter ...SelectFilter) (bool, error) {
	filter = append(filter, WithUserID(userID))

	return s.devices.Exists(filter...)
}

// Get returns a single device based on the provided filters for a specific user.
// It ensures that the filter includes the user's ID. If no device matches the
// criteria, it returns ErrNotFound. If more than one device matches, it returns
// ErrMoreThanOne.
func (s *Service) Get(userID string, filter ...SelectFilter) (models.Device, error) {
	filter = append(filter, WithUserID(userID))

	return s.devices.Get(filter...)
}

// GetByToken returns a device by token.
//
// This method is used to retrieve a device by its auth token. If the device
// does not exist, it returns ErrNotFound.
func (s *Service) GetByToken(token string) (models.Device, error) {
	device, err := s.cache.GetByToken(token)
	if err != nil {
		device, err = s.devices.Get(WithToken(token))
		if err != nil {
			return device, err
		}

		if err := s.cache.Set(device); err != nil {
			s.logger.Error("can't cache device", zap.String("device_id", device.ID), zap.Error(err))
		}
	}

	return device, nil
}

func (s *Service) UpdatePushToken(id string, token string) error {
	if err := s.cache.DeleteByID(id); err != nil {
		s.logger.Error("can't invalidate cache",
			zap.String("device_id", id),
			zap.Error(err),
		)
	}

	if err := s.devices.UpdatePushToken(id, token); err != nil {
		return err
	}

	return nil
}

func (s *Service) SetLastSeen(ctx context.Context, batch map[string]time.Time) error {
	if len(batch) == 0 {
		return nil
	}

	var multiErr error
	for deviceID, lastSeen := range batch {
		if err := ctx.Err(); err != nil {
			return errors.Join(err, multiErr)
		}
		if err := s.devices.SetLastSeen(ctx, deviceID, lastSeen); err != nil {
			multiErr = errors.Join(multiErr, fmt.Errorf("device %s: %w", deviceID, err))
			s.logger.Error("can't set last seen",
				zap.String("device_id", deviceID),
				zap.Time("last_seen", lastSeen),
				zap.Error(err),
			)
		}
	}
	return multiErr
}

// Remove removes devices for a specific user that match the provided filters.
// It ensures that the filter includes the user's ID.
func (s *Service) Remove(userID string, filter ...SelectFilter) error {
	filter = append(filter, WithUserID(userID))

	devices, err := s.devices.Select(filter...)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return nil
	}

	for _, device := range devices {
		if err := s.cache.DeleteByID(device.ID); err != nil {
			s.logger.Error("can't invalidate cache",
				zap.String("device_id", device.ID),
				zap.Error(err),
			)
		}
	}

	if err := s.devices.Remove(filter...); err != nil {
		return err
	}

	return nil
}
