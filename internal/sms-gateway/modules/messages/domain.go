package messages

import (
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
)

type MessageIn struct {
	ID           string
	Message      string
	PhoneNumbers []string
	IsEncrypted  bool

	SimNumber          *uint8
	WithDeliveryReport *bool
	TTL                *uint64
	ValidUntil         *time.Time
	Priority           smsgateway.MessagePriority
}

type MessageOut struct {
	MessageIn

	CreatedAt time.Time
}
