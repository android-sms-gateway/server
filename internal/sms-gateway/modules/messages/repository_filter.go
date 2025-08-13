package messages

import "time"

// MessagesOrder defines supported ordering for message selection.
// Valid values: "lifo" (default), "fifo".
type MessagesOrder string

const (
	// MessagesOrderLIFO orders messages newest-first within the same priority (default).
	MessagesOrderLIFO MessagesOrder = "lifo"
	// MessagesOrderFIFO orders messages oldest-first within the same priority.
	MessagesOrderFIFO MessagesOrder = "fifo"
)

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

	// OrderBy sets the retrieval order for pending messages.
	// Empty (zero) value defaults to "lifo".
	OrderBy MessagesOrder

	Limit  int
	Offset int
}
