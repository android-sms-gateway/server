package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/go-core-fx/cachefx/cache"
	"go.uber.org/zap"
)

type Service struct {
	cfg Config

	storage *Storage

	logger *zap.Logger
}

// NewService returns a new OTP service.
//
// It takes the configuration for the OTP service,
// a storage interface for storing the OTP codes,
// and a logger for logging events.
//
// It returns an error if the configuration is invalid,
// or if either the storage or logger is nil.
func NewService(cfg Config, storage *Storage, logger *zap.Logger) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if storage == nil {
		return nil, fmt.Errorf("%w: storage is required", ErrInitFailed)
	}

	if logger == nil {
		return nil, fmt.Errorf("%w: logger is required", ErrInitFailed)
	}

	return &Service{cfg: cfg, storage: storage, logger: logger}, nil
}

// Generate generates a new one-time user authorization code.
//
// It takes a context and the ID of the user for whom the code is being generated.
//
// It returns the generated code and an error. If the code can't be generated
// within the configured number of retries, it returns an error.
//
// The generated code is stored with a TTL of the configured duration.
func (s *Service) Generate(ctx context.Context, userID string) (*Code, error) {
	const maxValue = 1000000

	var code string
	var err error

	validUntil := time.Now().Add(s.cfg.TTL)

	for range s.cfg.Retries {
		num, rndErr := rand.Int(rand.Reader, big.NewInt(maxValue))
		if rndErr != nil {
			s.logger.Warn("failed to generate random number", zap.Error(rndErr))
			err = rndErr
			continue
		}

		code = fmt.Sprintf("%06d", num.Int64())

		if err = s.storage.SetOrFail(ctx, code, userID, cache.WithValidUntil(validUntil)); err != nil {
			s.logger.Warn("failed to store code", zap.Error(err))
			continue
		}

		err = nil
		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	return &Code{Code: code, ValidUntil: validUntil}, nil
}

// Validate validates a one-time user authorization code.
//
// It takes a context and the one-time code to be validated.
//
// It returns the user ID associated with the code, and an error.
// If the code is invalid, it returns ErrKeyNotFound.
// If there is an error while validating the code, it returns the error.
// If the code is valid, it deletes the code from the storage and returns the user ID.
func (s *Service) Validate(ctx context.Context, code string) (string, error) {
	return s.storage.GetAndDelete(ctx, code)
}
