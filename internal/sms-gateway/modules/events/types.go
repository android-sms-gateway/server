package events

import (
	"encoding/json"
	"fmt"

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
	data, err := json.Marshal(w)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	return data, nil
}

func (w *eventWrapper) deserialize(data []byte) error {
	if err := json.Unmarshal(data, w); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	return nil
}
