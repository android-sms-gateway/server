package messages

import "time"

type MessagesSelectFilter struct {
	ExtID     string
	UserID    string
	DeviceID  string
	StartDate time.Time
	EndDate   time.Time
	State     ProcessingState
}

type MessagesSelectOptions struct {
	WithRecipients bool
	WithDevice     bool
	WithStates     bool

	Limit  int
	Offset int
}
