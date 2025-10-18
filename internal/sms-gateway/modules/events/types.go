package events

import (
	"encoding/json"

	"github.com/android-sms-gateway/client-go/smsgateway"
)

type Event struct {
	EventType smsgateway.PushEventType `json:"event_type"`
	Data      map[string]string        `json:"data"`
}

func NewEvent(eventType smsgateway.PushEventType, data map[string]string) Event {
	return Event{
		EventType: eventType,
		Data:      data,
	}
}

type eventWrapper struct {
	UserID   string  `json:"user_id"`
	DeviceID *string `json:"device_id,omitempty"`
	Event    Event   `json:"event"`
}

func (w *eventWrapper) serialize() ([]byte, error) {
	return json.Marshal(w)
}

func (w *eventWrapper) deserialize(data []byte) error {
	return json.Unmarshal(data, w)
}
