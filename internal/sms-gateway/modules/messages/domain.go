package messages

import (
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
)

type MessageIn struct {
	ID string

	TextContent *TextMessageContent
	DataContent *DataMessageContent

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

type MessageStateIn struct {
	// Message ID
	ID string
	// State
	State ProcessingState
	// Recipients states
	Recipients []smsgateway.RecipientState
	// History of states
	States map[string]time.Time
}

type MessageStateOut struct {
	// Device ID
	DeviceID string
	// Hashed
	IsHashed bool
	// Encrypted
	IsEncrypted bool

	MessageStateIn
}
