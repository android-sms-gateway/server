package messages

import "time"

// Order defines supported ordering for message selection.
// Valid values: "lifo" (default), "fifo".
type Order string

const (
	// MessagesOrderLIFO orders messages newest-first within the same priority (default).
	MessagesOrderLIFO Order = "lifo"
	// MessagesOrderFIFO orders messages oldest-first within the same priority.
	MessagesOrderFIFO Order = "fifo"
)

type SelectFilter struct {
	ExtID     string
	UserID    string
	DeviceID  string
	StartDate time.Time
	EndDate   time.Time
	State     ProcessingState
}

func (f *SelectFilter) WithExtID(extID string) *SelectFilter {
	f.ExtID = extID
	return f
}

func (f *SelectFilter) WithUserID(userID string) *SelectFilter {
	f.UserID = userID
	return f
}

func (f *SelectFilter) WithDeviceID(deviceID string) *SelectFilter {
	f.DeviceID = deviceID
	return f
}

func (f *SelectFilter) WithDateRange(start, end time.Time) *SelectFilter {
	f.StartDate = start
	f.EndDate = end
	return f
}

func (f *SelectFilter) WithState(state ProcessingState) *SelectFilter {
	f.State = state
	return f
}

type SelectOptions struct {
	WithRecipients bool
	WithDevice     bool
	WithStates     bool

	// OrderBy sets the retrieval order for pending messages.
	// Empty (zero) value defaults to "lifo".
	OrderBy Order

	Limit  int
	Offset int
}

func (o *SelectOptions) WithLimit(limit int) *SelectOptions {
	o.Limit = limit
	return o
}

func (o *SelectOptions) WithOffset(offset int) *SelectOptions {
	o.Offset = offset
	return o
}

func (o *SelectOptions) WithOrderBy(order Order) *SelectOptions {
	o.OrderBy = order
	return o
}

func (o *SelectOptions) IncludeRecipients() *SelectOptions {
	o.WithRecipients = true
	return o
}

func (o *SelectOptions) IncludeDevice() *SelectOptions {
	o.WithDevice = true
	return o
}

func (o *SelectOptions) IncludeStates() *SelectOptions {
	o.WithStates = true
	return o
}
