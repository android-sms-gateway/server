package inbox

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/events"
	"go.uber.org/zap"
)

// Service provides business logic for inbox messages.
type Service struct {
	config Config

	eventsSvc *events.Service

	inbox *Repository

	logger *zap.Logger
}

// NewService creates a new Service.
func NewService(
	config Config,
	eventsSvc *events.Service,
	inbox *Repository,
	logger *zap.Logger,
) *Service {
	return &Service{
		config: config,

		eventsSvc: eventsSvc,

		inbox: inbox,

		logger: logger,
	}
}

func (s *Service) Refresh(
	userID string,
	deviceID *string,
	since, until time.Time,
	types []smsgateway.IncomingMessageType,
	webhookDelivery smsgateway.WebhookDelivery,
) error {
	event := events.NewMessagesExportRequestedEvent(since, until, types, webhookDelivery)

	if err := s.eventsSvc.Notify(userID, deviceID, event); err != nil {
		return fmt.Errorf("failed to notify device: %w", err)
	}

	return nil
}

// InsertBatch stores a batch of encrypted inbox messages.
// Returns ErrNotEncrypted if any message has IsEncrypted=false.
// Returns ErrEmptyBatch if the slice is empty.
func (s *Service) InsertBatch(ctx context.Context, deviceID string, msgs []MessageInput) error {
	if len(msgs) == 0 {
		return ErrEmptyBatch
	}

	for _, m := range msgs {
		if !m.IsEncrypted {
			return ErrNotEncrypted
		}
	}

	if err := s.inbox.InsertBatch(ctx, deviceID, msgs); err != nil {
		return fmt.Errorf("failed to insert inbox messages: %w", err)
	}

	return nil
}

// List returns inbox messages for a user.
func (s *Service) List(userID string, filter ListFilter, opts ListOptions) ([]Message, int64, error) {
	return s.inbox.list(userID, filter, opts)
}

// GetAttachment returns a single attachment, verifying user ownership of the parent message.
func (s *Service) GetAttachment(
	ctx context.Context,
	userID string,
	messageExtID string,
	partID int64,
) (*Attachment, error) {
	msg, err := s.inbox.findMessageByExtID(ctx, messageExtID, userID)
	if err != nil {
		return nil, err
	}

	return s.inbox.findAttachment(ctx, msg.ID, partID)
}
