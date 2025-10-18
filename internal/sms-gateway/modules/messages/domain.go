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
	ID         string                      `json:"id"`         // Message ID
	State      ProcessingState             `json:"state"`      // State
	Recipients []smsgateway.RecipientState `json:"recipients"` // Recipients states
	States     map[string]time.Time        `json:"states"`     // History of states
}

type MessageStateOut struct {
	MessageStateIn

	DeviceID    string `json:"device_id"`    // Device ID
	IsHashed    bool   `json:"is_hashed"`    // Hashed
	IsEncrypted bool   `json:"is_encrypted"` // Encrypted
}
