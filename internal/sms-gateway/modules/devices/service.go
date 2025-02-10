package devices

import (
	"context"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ServiceParams struct {
	fx.In

	Config Config

	Devices *repository

	IDGen db.IDGen

	Logger *zap.Logger
}

type Service struct {
	config Config

	devices *repository

	idGen db.IDGen

	logger *zap.Logger
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
	return s.devices.Get(WithToken(token))
}

func (s *Service) UpdateToken(deviceId string, token string) error {
	return s.devices.UpdateToken(deviceId, token)
}

func (s *Service) UpdateLastSeen(deviceId string) error {
	return s.devices.UpdateLastSeen(deviceId)
}

// Remove removes devices for a specific user that match the provided filters.
// It ensures that the filter includes the user's ID.
func (s *Service) Remove(userID string, filter ...SelectFilter) error {
	filter = append(filter, WithUserID(userID))

	return s.devices.Remove(filter...)
}

func (s *Service) Clean(ctx context.Context) error {
	n, err := s.devices.removeUnused(ctx, time.Now().Add(-s.config.UnusedLifetime))

	s.logger.Info("Cleaned unused devices", zap.Int64("count", n))
	return err
}

func NewService(params ServiceParams) *Service {
	return &Service{
		config:  params.Config,
		devices: params.Devices,
		idGen:   params.IDGen,
		logger:  params.Logger.Named("service"),
	}
}
