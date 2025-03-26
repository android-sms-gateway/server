package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/pkg/crypto"
	"github.com/capcom6/go-helpers/cache"
	"github.com/jaevor/go-nanoid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Config struct {
	Mode         Mode
	PrivateToken string
}

type Params struct {
	fx.In

	Config Config

	Users      *repository
	DevicesSvc *devices.Service

	Logger *zap.Logger
}

type Service struct {
	config Config

	users      *repository
	codesCache *cache.Cache[string]
	usersCache *cache.Cache[models.User]

	devicesSvc *devices.Service

	logger *zap.Logger

	idgen func() string
}

func New(params Params) *Service {
	idgen, _ := nanoid.Standard(21)

	return &Service{
		config:     params.Config,
		users:      params.Users,
		devicesSvc: params.DevicesSvc,
		logger:     params.Logger.Named("Service"),
		idgen:      idgen,

		codesCache: cache.New[string](cache.Config{}),
		usersCache: cache.New[models.User](cache.Config{TTL: 1 * time.Hour}),
	}
}

// GenerateUserCode generates a unique one-time user authorization code
func (s *Service) GenerateUserCode(userID string) (AuthCode, error) {
	var code string
	var err error

	b := make([]byte, 3)
	validUntil := time.Now().Add(codeTTL)
	for range 3 {
		if _, err = rand.Read(b); err != nil {
			continue
		}
		num := (int(b[0]) << 16) | (int(b[1]) << 8) | int(b[2])
		code = fmt.Sprintf("%06d", num%1000000)

		if err = s.codesCache.SetOrFail(code, userID, cache.WithValidUntil(validUntil)); err != nil {
			continue
		}

		break
	}

	if err != nil {
		return AuthCode{}, fmt.Errorf("can't generate code: %w", err)
	}

	return AuthCode{Code: code, ValidUntil: validUntil}, nil
}

func (s *Service) RegisterUser(login, password string) (models.User, error) {
	user := models.User{
		ID: login,
	}

	var err error
	if user.PasswordHash, err = crypto.MakeBCryptHash(password); err != nil {
		return user, fmt.Errorf("can't hash password: %w", err)
	}

	if err = s.users.Insert(&user); err != nil {
		return user, fmt.Errorf("can't create user")
	}

	return user, nil
}

func (s *Service) RegisterDevice(user models.User, name, pushToken *string) (models.Device, error) {
	device := models.Device{
		Name:      name,
		PushToken: pushToken,
	}

	return device, s.devicesSvc.Insert(user.ID, &device)
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

	return fmt.Errorf("invalid token")
}

func (s *Service) AuthorizeDevice(token string) (models.Device, error) {
	device, err := s.devicesSvc.GetByToken(token)
	if err != nil {
		return device, err
	}

	go func(id string) {
		if err := s.devicesSvc.UpdateLastSeen(id); err != nil {
			s.logger.Error("can't update last seen", zap.Error(err))
		}
	}(device.ID)

	device.LastSeen = time.Now()

	return device, nil
}

func (s *Service) AuthorizeUser(username, password string) (models.User, error) {
	hash := sha256.Sum256([]byte(username + password))
	cacheKey := hex.EncodeToString(hash[:])

	user, err := s.usersCache.Get(cacheKey)
	if err == nil {
		return user, nil
	}

	user, err = s.users.GetByLogin(username)
	if err != nil {
		return user, err
	}

	if err := crypto.CompareBCryptHash(user.PasswordHash, password); err != nil {
		return models.User{}, err
	}

	if err := s.usersCache.Set(cacheKey, user); err != nil {
		s.logger.Error("can't cache user", zap.Error(err))
	}

	return user, nil
}

// AuthorizeUserByCode authorizes a user by one-time code.
func (s *Service) AuthorizeUserByCode(code string) (models.User, error) {
	userID, err := s.codesCache.GetAndDelete(code)
	if err != nil {
		return models.User{}, err
	}

	user, err := s.users.GetByID(userID)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (s *Service) ChangePassword(userID string, currentPassword string, newPassword string) error {
	user, err := s.users.GetByLogin(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := crypto.CompareBCryptHash(user.PasswordHash, currentPassword); err != nil {
		return fmt.Errorf("current password is incorrect: %w", err)
	}

	newHash, err := crypto.MakeBCryptHash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.users.UpdatePassword(userID, newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate cache
	hash := sha256.Sum256([]byte(userID + currentPassword))
	cacheKey := hex.EncodeToString(hash[:])
	if err := s.usersCache.Delete(cacheKey); err != nil {
		s.logger.Error("can't invalidate user cache", zap.Error(err))
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
	s.codesCache.Cleanup()
	s.usersCache.Cleanup()
}
