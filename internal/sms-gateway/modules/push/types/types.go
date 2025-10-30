package types

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
)

type Message struct {
	Token string
	Event Event
}

type Event struct {
	Type smsgateway.PushEventType `json:"type"`
	Data map[string]string        `json:"data"`
}
