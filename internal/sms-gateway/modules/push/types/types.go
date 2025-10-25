package types

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
)

type Event struct {
	Type smsgateway.PushEventType `json:"type"`
	Data map[string]string        `json:"data"`
}
