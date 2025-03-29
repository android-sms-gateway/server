package messages

import "time"

type MessageIn struct {
	ID           string
	Message      string
	PhoneNumbers []string
	IsEncrypted  bool

	SimNumber          *uint8
	WithDeliveryReport *bool
	TTL                *uint64
	ValidUntil         *time.Time
}

type MessageOut struct {
	MessageIn

	CreatedAt time.Time
}
