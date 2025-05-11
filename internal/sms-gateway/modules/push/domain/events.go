package domain

import (
	"encoding/json"

	"github.com/android-sms-gateway/client-go/smsgateway"
)

type Event struct {
	event smsgateway.PushEventType
	data  map[string]string
}

func (e *Event) Event() smsgateway.PushEventType {
	return e.event
}

func (e *Event) Data() map[string]string {
	return e.data
}

func (e *Event) Map() map[string]string {
	json, _ := json.Marshal(e.data)

	return map[string]string{
		"event": string(e.event),
		"data":  string(json),
	}
}

func NewEvent(event smsgateway.PushEventType, data map[string]string) *Event {
	return &Event{
		event: event,
		data:  data,
	}
}
