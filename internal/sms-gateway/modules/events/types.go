package events

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
)

type Event struct {
	eventType smsgateway.PushEventType
	data      map[string]string
}

func NewEvent(eventType smsgateway.PushEventType, data map[string]string) *Event {
	return &Event{
		eventType: eventType,
		data:      data,
	}
}

type eventWrapper struct {
	UserID   string
	DeviceID *string
	Event    *Event
}
