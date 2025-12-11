package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/online"
	"github.com/android-sms-gateway/server/internal/sms-gateway/otp"
	"github.com/android-sms-gateway/server/internal/sms-gateway/users"
	"go.uber.org/zap"
)

type Config struct {
	Mode         Mode
	PrivateToken string
}

type Service struct {
	config Config

	usersSvc   *users.Service
	otpSvc     *otp.Service
	devicesSvc *devices.Service
	onlineSvc  online.Service

	logger *zap.Logger
}

func New(
	config Config,
	usersSvc *users.Service,
	otpSvc *otp.Service,
	devicesSvc *devices.Service,
	onlineSvc online.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		config: config,

		usersSvc: usersSvc,

		otpSvc:     otpSvc,
		devicesSvc: devicesSvc,
		onlineSvc:  onlineSvc,

		logger: logger,
	}
}

// GenerateUserCode generates a unique one-time user authorization code.
func (s *Service) GenerateUserCode(ctx context.Context, userID string) (*otp.Code, error) {
	code, err := s.otpSvc.Generate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	return code, nil
}

func (s *Service) RegisterDevice(userID string, name, pushToken *string) (*models.Device, error) {
	device := models.NewDevice(
		name,
		pushToken,
	)

	if err := s.devicesSvc.Insert(userID, device); err != nil {
		return device, fmt.Errorf("failed to create device: %w", err)
	}

	return device, nil
}

func (s *Service) IsPublic() bool {
	return s.config.Mode == ModePublic
}

func (s *Service) AuthorizeRegistration(token string) error {
	if s.IsPublic() {
		return nil
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(s.config.PrivateToken)) == 1 {
		return nil
	}

	return ErrAuthorizationFailed
}

func (s *Service) AuthorizeDevice(token string) (models.Device, error) {
	device, err := s.devicesSvc.GetByToken(token)
	if err != nil {
		return device, fmt.Errorf("%w: %w", ErrAuthorizationFailed, err)
	}

	go func(id string) {
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		s.onlineSvc.SetOnline(ctx, id)
	}(device.ID)

	device.LastSeen = time.Now()

	return device, nil
}

// AuthorizeUserByCode authorizes a user by one-time code.
func (s *Service) AuthorizeUserByCode(ctx context.Context, code string) (*users.User, error) {
	userID, err := s.otpSvc.Validate(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to validate code: %w", err)
	}

	user, err := s.usersSvc.GetByUsername(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
