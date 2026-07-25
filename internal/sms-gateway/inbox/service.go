package inbox

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
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
	triggerWebhooks *bool,
) error {
	event := events.NewMessagesExportRequestedEvent(since, until, types, triggerWebhooks)

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

	result := make([]*messageModel, 0, len(msgs))
	for _, m := range msgs {
		if !m.IsEncrypted {
			return ErrNotEncrypted
		}

		atts := make([]attachmentModel, 0, len(m.Attachments))
		for _, a := range m.Attachments {
			if !a.IsEncrypted {
				return ErrNotEncrypted
			}
			atts = append(atts, attachmentModel{
				ID:          0,
				MessageID:   0,
				PartID:      a.PartID,
				ContentType: a.ContentType,
				Name:        a.Name,
				Size:        a.Size,
				Data:        a.Data,
			})
		}

		result = append(result, &messageModel{
			TimedModel:  models.TimedModel{CreatedAt: m.CreatedAt, UpdatedAt: m.CreatedAt},
			ID:          0,
			ExtID:       m.ID,
			DeviceID:    deviceID,
			Type:        m.Type,
			Sender:      m.Sender,
			Recipient:   m.Recipient,
			SimNumber:   m.SimNumber,
			Content:     m.Content,
			IsEncrypted: true,
			Attachments: atts,
		})
	}

	if err := s.inbox.InsertBatch(ctx, result); err != nil {
		return fmt.Errorf("failed to insert inbox messages: %w", err)
	}

	return nil
}

// List returns inbox messages for a user.
func (s *Service) List(userID string, filter ListFilter, opts ListOptions) ([]Message, int64, error) {
	filter.UserID = userID

	messages, total, err := s.inbox.list(filter, opts)
	if err != nil {
		return nil, 0, err
	}

	result := make([]Message, len(messages))
	for i, m := range messages {
		atts := make([]Attachment, len(m.Attachments))
		for j, a := range m.Attachments {
			atts[j] = Attachment{
				ID:          a.ID,
				PartID:      a.PartID,
				ContentType: a.ContentType,
				Name:        a.Name,
				Size:        a.Size,
				Data:        a.Data,
			}
		}
		result[i] = Message{
			ID:          m.ID,
			ExtID:       m.ExtID,
			DeviceID:    m.DeviceID,
			Type:        m.Type,
			Sender:      m.Sender,
			Recipient:   m.Recipient,
			SimNumber:   m.SimNumber,
			Content:     m.Content,
			IsEncrypted: m.IsEncrypted,
			CreatedAt:   m.CreatedAt,
			Attachments: atts,
		}
	}

	return result, total, nil
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

	att, err := s.inbox.findAttachment(ctx, msg.ID, partID)
	if err != nil {
		return nil, err
	}

	return &Attachment{
		ID:          att.ID,
		PartID:      att.PartID,
		ContentType: att.ContentType,
		Name:        att.Name,
		Size:        att.Size,
		Data:        att.Data,
	}, nil
}
