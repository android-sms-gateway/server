package types

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
)

type Event struct {
	Type smsgateway.PushEventType
	Data map[string]string
}
