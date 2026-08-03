package otp_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/otp"
	"github.com/go-core-fx/cachefx/cache"
	"github.com/go-playground/assert/v2"
	"go.uber.org/zap"
)

func TestNewService(t *testing.T) {
	storage := otp.NewStorage(cache.NewMemory(time.Hour))
	logger := zap.NewNop()

	tests := []struct {
		name    string
		cfg     otp.Config
		storage *otp.Storage
		logger  *zap.Logger
		wantErr error
	}{
		{
			name: "valid",
			cfg: otp.Config{
				Enabled: true,
				Length:  8,
				TTL:     time.Minute,
				Retries: 3,
			},
			storage: storage,
			logger:  logger,
		},
		{
			name: "nil storage",
			cfg: otp.Config{
				Enabled: true,
				Length:  8,
				TTL:     time.Minute,
				Retries: 3,
			},
			storage: nil,
			logger:  logger,
			wantErr: otp.ErrInitFailed,
		},
		{
			name: "nil logger",
			cfg: otp.Config{
				Enabled: true,
				Length:  8,
				TTL:     time.Minute,
				Retries: 3,
			},
			storage: storage,
			logger:  nil,
			wantErr: otp.ErrInitFailed,
		},
		{
			name: "invalid length",
			cfg: otp.Config{
				Enabled: true,
				Length:  3,
				TTL:     time.Minute,
				Retries: 3,
			},
			storage: storage,
			logger:  logger,
			wantErr: otp.ErrInvalidConfig,
		},
		{
			name: "zero ttl",
			cfg: otp.Config{
				Enabled: true,
				Length:  8,
				TTL:     0,
				Retries: 3,
			},
			storage: storage,
			logger:  logger,
			wantErr: otp.ErrInvalidConfig,
		},
		{
			name: "zero retries",
			cfg: otp.Config{
				Enabled: true,
				Length:  8,
				TTL:     time.Minute,
				Retries: 0,
			},
			storage: storage,
			logger:  logger,
			wantErr: otp.ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := otp.NewService(tt.cfg, tt.storage, tt.logger)
			if tt.wantErr != nil {
				assert.Equal(t, true, errors.Is(err, tt.wantErr))
				assert.Equal(t, (*otp.Service)(nil), svc)
			} else {
				assert.Equal(t, nil, err)
				assert.Equal(t, true, svc != nil)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	ctx := context.Background()

	t.Run("disabled", func(t *testing.T) {
		svc, err := otp.NewService(
			otp.Config{Enabled: false, Length: 8, TTL: time.Minute, Retries: 3},
			otp.NewStorage(cache.NewMemory(time.Hour)),
			zap.NewNop(),
		)
		assert.Equal(t, nil, err)

		code, err := svc.Generate(ctx, "user1")
		assert.Equal(t, otp.ErrDisabled, err)
		assert.Equal(t, (*otp.Code)(nil), code)
	})

	t.Run("enabled code length and alphabet", func(t *testing.T) {
		const length = 10
		svc, err := otp.NewService(
			otp.Config{Enabled: true, Length: length, TTL: time.Minute, Retries: 3},
			otp.NewStorage(cache.NewMemory(time.Hour)),
			zap.NewNop(),
		)
		assert.Equal(t, nil, err)

		code, err := svc.Generate(ctx, "user1")
		assert.Equal(t, nil, err)
		assert.Equal(t, true, code != nil)
		assert.Equal(t, length, len(code.Code))

		for _, ch := range code.Code {
			assert.Equal(t, true, strings.ContainsRune(otp.Alphabet(), ch))
		}
	})
}

func TestValidate(t *testing.T) {
	ctx := context.Background()

	t.Run("disabled", func(t *testing.T) {
		svc, err := otp.NewService(
			otp.Config{Enabled: false, Length: 8, TTL: time.Minute, Retries: 3},
			otp.NewStorage(cache.NewMemory(time.Hour)),
			zap.NewNop(),
		)
		assert.Equal(t, nil, err)

		userID, err := svc.Validate(ctx, "abc123")
		assert.Equal(t, otp.ErrDisabled, err)
		assert.Equal(t, "", userID)
	})

	t.Run("valid code", func(t *testing.T) {
		svc, err := otp.NewService(
			otp.Config{Enabled: true, Length: 8, TTL: time.Minute, Retries: 3},
			otp.NewStorage(cache.NewMemory(time.Hour)),
			zap.NewNop(),
		)
		assert.Equal(t, nil, err)

		code, err := svc.Generate(ctx, "user1")
		assert.Equal(t, nil, err)

		userID, err := svc.Validate(ctx, code.Code)
		assert.Equal(t, nil, err)
		assert.Equal(t, "user1", userID)
	})

	t.Run("mixed case lookup", func(t *testing.T) {
		svc, err := otp.NewService(
			otp.Config{Enabled: true, Length: 8, TTL: time.Minute, Retries: 3},
			otp.NewStorage(cache.NewMemory(time.Hour)),
			zap.NewNop(),
		)
		assert.Equal(t, nil, err)

		code, err := svc.Generate(ctx, "user1")
		assert.Equal(t, nil, err)

		upper := strings.ToUpper(code.Code)
		userID, err := svc.Validate(ctx, upper)
		assert.Equal(t, nil, err)
		assert.Equal(t, "user1", userID)
	})

	t.Run("invalid code", func(t *testing.T) {
		svc, err := otp.NewService(
			otp.Config{Enabled: true, Length: 8, TTL: time.Minute, Retries: 3},
			otp.NewStorage(cache.NewMemory(time.Hour)),
			zap.NewNop(),
		)
		assert.Equal(t, nil, err)

		userID, err := svc.Validate(ctx, "nonexistent")
		assert.Equal(t, true, err != nil)
		assert.Equal(t, "", userID)
	})
}
