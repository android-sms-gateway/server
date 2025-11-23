package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/online"
	"github.com/android-sms-gateway/server/internal/sms-gateway/users"
	"github.com/capcom6/go-helpers/cache"
	"go.uber.org/zap"
)

type Config struct {
	Mode         Mode
	PrivateToken string
}

type Service struct {
	config Config

	usersSvc   *users.Service
	devicesSvc *devices.Service
	onlineSvc  online.Service

	logger *zap.Logger

	codesCache *cache.Cache[string]
}

func New(
	config Config,
	usersSvc *users.Service,
	devicesSvc *devices.Service,
	onlineSvc online.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		config: config,

		usersSvc:   usersSvc,
		devicesSvc: devicesSvc,
		onlineSvc:  onlineSvc,

		logger: logger,

		codesCache: cache.New[string](cache.Config{TTL: codeTTL}),
	}
}

// GenerateUserCode generates a unique one-time user authorization code.
func (s *Service) GenerateUserCode(userID string) (OneTimeCode, error) {
	var code string
	var err error

	const bytesLen = 3
	const maxCode = 1000000
	b := make([]byte, bytesLen)
	validUntil := time.Now().Add(codeTTL)
	for range 3 {
		if _, err = rand.Read(b); err != nil {
			continue
		}
		num := (int(b[0]) << 16) | (int(b[1]) << 8) | int(b[2]) //nolint:mnd //bitshift
		code = fmt.Sprintf("%06d", num%maxCode)

		if err = s.codesCache.SetOrFail(code, userID, cache.WithValidUntil(validUntil)); err != nil {
			continue
		}

		break
	}

	if err != nil {
		return OneTimeCode{}, fmt.Errorf("failed to generate code: %w", err)
	}

	return OneTimeCode{Code: code, ValidUntil: validUntil}, nil
}

func (s *Service) RegisterDevice(user users.User, name, pushToken *string) (*models.Device, error) {
	device := models.NewDevice(
		name,
		pushToken,
	)

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

// AuthorizeUserByCode authorizes a user by one-time code.
func (s *Service) AuthorizeUserByCode(code string) (*users.User, error) {
	userID, err := s.codesCache.GetAndDelete(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by code: %w", err)
	}

	user, err := s.usersSvc.GetByUsername(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
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
	s.codesCache.Cleanup()
}
