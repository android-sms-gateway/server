package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/android-sms-gateway/server/pkg/cache"
	"github.com/android-sms-gateway/server/pkg/crypto"
	"go.uber.org/zap"
)

type Service struct {
	users *repository

	cache *loginCache

	logger *zap.Logger
}

func NewService(
	users *repository,
	cache *loginCache,
	logger *zap.Logger,
) *Service {
	return &Service{
		users: users,

		cache: cache,

		logger: logger,
	}
}

func (s *Service) Create(username, password string) (*User, error) {
	exists, err := s.users.Exists(username)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("%w: %s", ErrExists, username)
	}

	passwordHash, err := crypto.MakeBCryptHash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &userModel{
		ID:           username,
		PasswordHash: passwordHash,
	}

	if err := s.users.Insert(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser(user), nil
}

func (s *Service) GetByUsername(username string) (*User, error) {
	user, err := s.users.GetByID(username)
	if err != nil {
		return nil, err
	}

	return newUser(user), nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*User, error) {
	cachedUser, err := s.cache.Get(ctx, username, password)
	if err == nil {
		return cachedUser, nil
	} else if !errors.Is(err, cache.ErrKeyNotFound) {
		s.logger.Warn("failed to get user from cache", zap.String("username", username), zap.Error(err))
	}

	user, err := s.users.GetByID(username)
	if err != nil {
		return nil, err
	}

	if err := crypto.CompareBCryptHash(user.PasswordHash, password); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	loggedInUser := newUser(user)
	if err := s.cache.Set(ctx, username, password, *loggedInUser); err != nil {
		s.logger.Error("failed to cache user", zap.String("username", username), zap.Error(err))
	}

	return loggedInUser, nil
}

func (s *Service) ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error {
	_, err := s.Login(ctx, username, currentPassword)
	if err != nil {
		return err
	}

	if err := s.cache.Delete(ctx, username, currentPassword); err != nil {
		return err
	}

	passwordHash, err := crypto.MakeBCryptHash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.users.UpdatePassword(username, passwordHash)
}
