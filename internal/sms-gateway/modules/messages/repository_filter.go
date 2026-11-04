package messages

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Order defines supported ordering for message selection.
// Valid values: "lifo" (default), "fifo".
type Order string

const (
	// MessagesOrderLIFO orders messages newest-first within the same priority (default).
	MessagesOrderLIFO Order = "lifo"
	// MessagesOrderFIFO orders messages oldest-first within the same priority.
	MessagesOrderFIFO Order = "fifo"
)

// SortField defines the allowed column sort directions.
// Only values defined here may be used in SelectOptions.SortField.
type SortField string

const (
	// SortFieldNone is the zero value (no explicit sort; falls through to OrderBy/default).
	SortFieldNone SortField = ""
	// SortFieldCreatedAtAsc sorts by created_at ascending.
	SortFieldCreatedAtAsc SortField = "created_at_asc"
	// SortFieldCreatedAtDesc sorts by created_at descending.
	SortFieldCreatedAtDesc SortField = "created_at_desc"
)

type SelectFilter struct {
	ExtID     string
	UserID    string
	DeviceID  string
	StartDate time.Time
	EndDate   time.Time
	State     []ProcessingState
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
	f.State = append(f.State, state)
	return f
}

// hash returns the identity of the row set this filter selects, as a
// cache key: every field that changes the counted rows is hashed, while
// pagination is deliberately absent because a total does not depend on the
// requested page.
func (f *SelectFilter) hash() string {
	key := f.UserID + "\x00" + f.ExtID + "\x00" + f.DeviceID + "\x00" +
		strconv.FormatInt(f.StartDate.UnixNano(), 10) + "\x00" +
		strconv.FormatInt(f.EndDate.UnixNano(), 10)

	if len(f.State) > 0 {
		states := make([]string, 0, len(f.State))
		for _, state := range f.State {
			states = append(states, string(state))
		}
		key += "\x00" + strings.Join(states, ",")
	}

	sum := sha256.Sum256([]byte(key))

	return hex.EncodeToString(sum[:])
}

// apply adds the WHERE clauses shared by every message selection to query.
func (f *SelectFilter) apply(query *gorm.DB) *gorm.DB {
	// Apply date range filter
	if !f.StartDate.IsZero() {
		query = query.Where("messages.created_at >= ?", f.StartDate)
	}
	if !f.EndDate.IsZero() {
		query = query.Where("messages.created_at < ?", f.EndDate)
	}

	// Apply ID filter
	if f.ExtID != "" {
		query = query.Where("messages.ext_id = ?", f.ExtID)
	}

	// Apply user filter
	if f.UserID != "" {
		query = query.Where("messages.user_id = ?", f.UserID)
	}

	// Apply state filter
	if len(f.State) > 0 {
		query = query.Where("messages.state IN ?", f.State)
	}

	// Apply device filter
	if f.DeviceID != "" {
		query = query.Where("messages.device_id = ?", f.DeviceID)
	}

	return query
}

type SelectOptions struct {
	WithRecipients bool
	WithDevice     bool
	WithStates     bool
	WithContent    bool

	// OrderBy sets the retrieval order for pending messages.
	// Empty (zero) value defaults to "lifo".
	OrderBy Order

	// SortField selects a column sort direction.
	// When non-zero, it overrides OrderBy.
	SortField SortField

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

func (o *SelectOptions) IncludeContent() *SelectOptions {
	o.WithContent = true
	return o
}
