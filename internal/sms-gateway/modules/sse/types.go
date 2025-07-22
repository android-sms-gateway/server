package sse

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
)

type Event struct {
	Type smsgateway.PushEventType `json:"event"`
	Data map[string]string        `json:"data"`
}
