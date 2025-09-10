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

	logger  *zap.Logger
	metrics *metrics
}

func New(devicesSvc *devices.Service, cache cache.Cache, logger *zap.Logger, metrics *metrics) Service {
	return &service{
		devicesSvc: devicesSvc,

		cache: cache,

		logger:  logger,
		metrics: metrics,
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
			s.logger.Debug("Persisting online status")
			if err := s.persist(ctx); err != nil {
				s.logger.Error("Can't persist online status", zap.Error(err))
			}
		}
	}
}

func (s *service) SetOnline(ctx context.Context, deviceID string) {
	dt := time.Now().UTC().Format(time.RFC3339)

	s.logger.Debug("Setting online status", zap.String("device_id", deviceID), zap.String("last_seen", dt))

	var err error
	s.metrics.ObserveCacheLatency(func() {
		if err = s.cache.Set(ctx, deviceID, dt); err != nil {
			s.metrics.IncrementCacheOperation(operationSet, statusError)
			s.logger.Error("Can't set online status", zap.String("device_id", deviceID), zap.Error(err))
			s.metrics.IncrementStatusSet(false)
		}
	})

	if err != nil {
		return
	}

	s.metrics.IncrementCacheOperation(operationSet, statusSuccess)
	s.logger.Debug("Online status set", zap.String("device_id", deviceID))
	s.metrics.IncrementStatusSet(true)
}

func (s *service) persist(ctx context.Context) error {
	var drainErr, persistErr error

	s.metrics.ObservePersistenceLatency(func() {
		items, err := s.cache.Drain(ctx)
		if err != nil {
			drainErr = fmt.Errorf("can't drain cache: %w", err)
			s.metrics.IncrementCacheOperation(operationDrain, statusError)
			return
		}
		s.metrics.IncrementCacheOperation(operationDrain, statusSuccess)
		s.metrics.SetBatchSize(len(items))

		if len(items) == 0 {
			s.logger.Debug("No online statuses to persist")
			return
		}
		s.logger.Debug("Drained cache", zap.Int("count", len(items)))

		timestamps := maps.MapValues(items, func(v string) time.Time {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				s.logger.Warn("Can't parse last seen", zap.String("last_seen", v), zap.Error(err))
				return time.Now().UTC()
			}

			return t
		})

		s.logger.Debug("Parsed last seen timestamps", zap.Int("count", len(timestamps)))

		if err := s.devicesSvc.SetLastSeen(ctx, timestamps); err != nil {
			persistErr = fmt.Errorf("can't set last seen: %w", err)
			s.metrics.IncrementPersistenceError()
			return
		}

		s.logger.Info("Set last seen", zap.Int("count", len(timestamps)))
	})

	if drainErr != nil {
		return drainErr
	}

	if persistErr != nil {
		return persistErr
	}

	return nil
}
