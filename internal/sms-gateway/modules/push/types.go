package push

import (
	"context"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/domain"
)

type Mode string
type Event = domain.Event

const (
	ModeFCM      Mode = "fcm"
	ModeUpstream Mode = "upstream"
)

type client interface {
	Open(ctx context.Context) error
	Send(ctx context.Context, messages map[string]domain.Event) error
	Close(ctx context.Context) error
}

func NewMessageEnqueuedEvent() *domain.Event {
	return domain.NewEvent(smsgateway.PushMessageEnqueued, nil)
}

func NewWebhooksUpdatedEvent() *domain.Event {
	return domain.NewEvent(smsgateway.PushWebhooksUpdated, nil)
}

func NewMessagesExportRequestedEvent(since, until time.Time) *domain.Event {
	return domain.NewEvent(
		smsgateway.PushMessagesExportRequested,
		map[string]string{
			"since": since.Format(time.RFC3339),
			"until": until.Format(time.RFC3339),
		},
	)
}
