package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/pkg/cache"
	"github.com/samber/lo"
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
	const maxValue = 999999
	const bytesLength = 3

	var code string
	var err error

	b := make([]byte, bytesLength)
	validUntil := time.Now().Add(s.cfg.TTL)

	for range s.cfg.Retries {
		if _, err = rand.Read(b); err != nil {
			s.logger.Warn("failed to read random bytes", zap.Error(err))
			continue
		}

		num := lo.Reduce(
			b,
			func(agg uint64, item byte, _ int) uint64 {
				return (agg << 8) | uint64(item) //nolint:mnd // bits in byte
			},
			0,
		)
		code = fmt.Sprintf("%06d", num%(maxValue+1))

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
