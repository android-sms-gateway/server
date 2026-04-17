package inbox

import (
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/events"
	"go.uber.org/zap"
)

type Service struct {
	eventsSvc *events.Service

	logger *zap.Logger
}

func New(eventsSvc *events.Service, logger *zap.Logger) *Service {
	return &Service{
		eventsSvc: eventsSvc,

		logger: logger,
	}
}

func (s *Service) Refresh(userID string, deviceID *string, since, until time.Time) error {
	event := events.NewMessagesExportRequestedEvent(since, until)

	if err := s.eventsSvc.Notify(userID, deviceID, event); err != nil {
		return fmt.Errorf("failed to notify device: %w", err)
	}

	return nil
}
