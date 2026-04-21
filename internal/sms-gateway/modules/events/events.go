package events

import (
	"strconv"
	"strings"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
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
	triggerWebhooks *bool,
) Event {
	data := map[string]string{
		"since": since.Format(time.RFC3339),
		"until": until.Format(time.RFC3339),
	}
	if len(types) > 0 {
		str := strings.Builder{}
		str.WriteString("[")
		for i, t := range types {
			if i > 0 {
				str.WriteString(",")
			}
			str.WriteString(strconv.Quote(string(t)))
		}
		str.WriteString("]")
		data["messageTypes"] = str.String()
	}
	if triggerWebhooks != nil {
		data["triggerWebhooks"] = strconv.FormatBool(*triggerWebhooks)
	}

	return NewEvent(
		smsgateway.PushMessagesExportRequested,
		data,
	)
}

func NewSettingsUpdatedEvent() Event {
	return NewEvent(smsgateway.PushSettingsUpdated, nil)
}
