package events

import (
	"strconv"
	"strings"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/samber/lo"
)

func NewMessageEnqueuedEvent() Event {
	return NewEvent(smsgateway.PushMessageEnqueued, nil)
}

func NewWebhooksUpdatedEvent() Event {
	return NewEvent(smsgateway.PushWebhooksUpdated, nil)
}

func NewMessagesExportRequestedEvent(
	since, until time.Time,
	types []smsgateway.IncomingMessageType,
	webhookDelivery smsgateway.WebhookDelivery,
) Event {
	data := map[string]string{
		"since": since.Format(time.RFC3339),
		"until": until.Format(time.RFC3339),
	}
	if len(types) > 0 {
		data["messageTypes"] = strings.Join(
			lo.Map(types, func(item smsgateway.IncomingMessageType, _ int) string { return string(item) }),
			",",
		)
	}
	data["webhookDelivery"] = string(webhookDelivery)
	data["triggerWebhooks"] = strconv.FormatBool(webhookDelivery != smsgateway.WebhookDeliveryDisabled)

	return NewEvent(
		smsgateway.PushMessagesExportRequested,
		data,
	)
}

func NewMessageCancelledEvent(messageID string) Event {
	return NewEvent(smsgateway.PushMessageCancelled, map[string]string{
		"messageId": messageID,
	})
}

func NewSettingsUpdatedEvent() Event {
	return NewEvent(smsgateway.PushSettingsUpdated, nil)
}
