package messages

import (
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
)

type TextContent = smsgateway.TextMessage
type DataContent = smsgateway.DataMessage
type MultimediaContent = smsgateway.MmsMessage
type HashedMessageContent = smsgateway.HashedMessage

type MessageContent struct {
	TextContent *TextContent       `json:"textContent,omitempty"`
	DataContent *DataContent       `json:"dataContent,omitempty"`
	MmsContent  *MultimediaContent `json:"mmsContent,omitempty"`
}

type MessageStateContent struct {
	MessageContent

	HashedContent *HashedMessageContent `json:"hashedContent,omitempty"`
}

type MessageInput struct {
	MessageContent

	ID string

	PhoneNumbers []string
	IsEncrypted  bool

	SimNumber          *uint8
	WithDeliveryReport *bool
	TTL                *uint64
	ValidUntil         *time.Time
	ScheduleAt         *time.Time
	Priority           smsgateway.MessagePriority
}

type Message struct {
	MessageInput

	State     ProcessingState
	CreatedAt time.Time
}

type MessageStateInput struct {
	ID         string                      `json:"id"`         // Message ID
	State      ProcessingState             `json:"state"`      // State
	Recipients []smsgateway.RecipientState `json:"recipients"` // Recipients states
	States     map[string]time.Time        `json:"states"`     // History of states
}

type MessageState struct {
	MessageStateInput
	MessageStateContent

	DeviceID    string `json:"deviceId"`    // Device ID
	IsHashed    bool   `json:"isHashed"`    // Hashed
	IsEncrypted bool   `json:"isEncrypted"` // Encrypted
}
