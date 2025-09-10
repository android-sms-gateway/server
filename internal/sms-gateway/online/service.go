package online

import (
	"context"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/pkg/cache"
	"go.uber.org/zap"
)

type Service struct {
	devicesSvc *devices.Service

	cache cache.Cache

	logger *zap.Logger
}

func New(devicesSvc *devices.Service, cache cache.Cache, logger *zap.Logger) *Service {
	return &Service{
		devicesSvc: devicesSvc,

		cache: cache,

		logger: logger,
	}
}

func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.logger.Info("Checking online status")
		}
	}
}
