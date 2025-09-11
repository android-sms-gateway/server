package online

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/pkg/cache"
	"github.com/capcom6/go-helpers/maps"
	"go.uber.org/zap"
)

type Service interface {
	Run(ctx context.Context)
	SetOnline(ctx context.Context, deviceID string)
}

type service struct {
	devicesSvc *devices.Service

	cache cache.Cache

	logger *zap.Logger
}

func New(devicesSvc *devices.Service, cache cache.Cache, logger *zap.Logger) Service {
	return &service{
		devicesSvc: devicesSvc,

		cache: cache,

		logger: logger,
	}
}

func (s *service) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.logger.Info("Persisting online status")
			if err := s.persist(ctx); err != nil {
				s.logger.Error("Can't persist online status", zap.Error(err))
			} else {
				s.logger.Info("Online status persisted")
			}
		}
	}
}

func (s *service) SetOnline(ctx context.Context, deviceID string) {
	dt := time.Now().UTC().Format(time.RFC3339)

	s.logger.Info("Setting online status", zap.String("device_id", deviceID), zap.String("last_seen", dt))

	if err := s.cache.Set(ctx, deviceID, dt); err != nil {
		s.logger.Error("Can't set online status", zap.String("device_id", deviceID), zap.Error(err))
		return
	}

	s.logger.Info("Online status set", zap.String("device_id", deviceID))
}

func (s *service) persist(ctx context.Context) error {
	items, err := s.cache.Drain(ctx)
	if err != nil {
		return fmt.Errorf("can't drain cache: %w", err)
	}

	s.logger.Info("Drained cache", zap.Int("count", len(items)))

	timestamps := maps.MapValues(items, func(v string) time.Time {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			s.logger.Error("Can't parse last seen", zap.String("last_seen", v), zap.Error(err))
			return time.Now().UTC()
		}

		return t
	})

	s.logger.Info("Parsed last seen timestamps", zap.Int("count", len(timestamps)))

	if err := s.devicesSvc.SetLastSeen(ctx, timestamps); err != nil {
		return fmt.Errorf("can't set last seen: %w", err)
	}

	s.logger.Info("Set last seen")

	return nil
}
