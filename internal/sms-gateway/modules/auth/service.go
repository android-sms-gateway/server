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
	"github.com/android-sms-gateway/server/pkg/crypto"
	"go.uber.org/zap"
)

type Config struct {
	Mode         Mode
	PrivateToken string
}

type Service struct {
	config Config

	users      *repository
	usersCache *usersCache

	otpSvc     *otp.Service
	devicesSvc *devices.Service
	onlineSvc  online.Service

	logger *zap.Logger
}

func New(
	config Config,
	users *repository,
	otpSvc *otp.Service,
	devicesSvc *devices.Service,
	onlineSvc online.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		config: config,

		users: users,

		otpSvc:     otpSvc,
		devicesSvc: devicesSvc,
		onlineSvc:  onlineSvc,

		logger: logger,

		usersCache: newUsersCache(),
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

func (s *Service) RegisterUser(login, password string) (*models.User, error) {
	passwordHash, err := crypto.MakeBCryptHash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.NewUser(login, passwordHash)
	if err = s.users.Insert(user); err != nil {
		return user, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *Service) RegisterDevice(user *models.User, name, pushToken *string) (*models.Device, error) {
	device := models.NewDevice(name, pushToken)

	if err := s.devicesSvc.Insert(user.ID, device); err != nil {
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

func (s *Service) AuthorizeUser(username, password string) (*models.User, error) {
	if user, err := s.usersCache.Get(username, password); err == nil {
		return &user, nil
	}

	user, err := s.users.GetByLogin(username)
	if err != nil {
		return user, err
	}

	if cmpErr := crypto.CompareBCryptHash(user.PasswordHash, password); cmpErr != nil {
		return nil, fmt.Errorf("password is incorrect: %w", cmpErr)
	}

	if setErr := s.usersCache.Set(username, password, *user); setErr != nil {
		s.logger.Error("failed to cache user", zap.Error(setErr))
	}

	return user, nil
}

// AuthorizeUserByCode authorizes a user by one-time code.
func (s *Service) AuthorizeUserByCode(ctx context.Context, code string) (*models.User, error) {
	userID, err := s.otpSvc.Validate(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to validate code: %w", err)
	}

	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *Service) ChangePassword(userID string, currentPassword string, newPassword string) error {
	user, err := s.users.GetByLogin(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if hashErr := crypto.CompareBCryptHash(user.PasswordHash, currentPassword); hashErr != nil {
		return fmt.Errorf("current password is incorrect: %w", hashErr)
	}

	newHash, err := crypto.MakeBCryptHash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if updErr := s.users.UpdatePassword(userID, newHash); updErr != nil {
		return fmt.Errorf("failed to update password: %w", updErr)
	}

	// Invalidate cache
	if delErr := s.usersCache.Delete(userID, currentPassword); delErr != nil {
		s.logger.Error("failed to invalidate user cache", zap.Error(delErr))
	}

	return nil
}

// Run starts a ticker that triggers the clean function every hour.
// It runs indefinitely until the provided context is canceled.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.clean(ctx)
		}
	}
}

func (s *Service) clean(_ context.Context) {
	s.usersCache.Cleanup()
}
