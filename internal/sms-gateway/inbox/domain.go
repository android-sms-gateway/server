package inbox

import (
	"time"

	"gorm.io/gorm"
)

// MessageType represents the type of inbox message.
type MessageType string

const (
	MessageTypeSMS           MessageType = "SMS"
	MessageTypeDataSMS       MessageType = "DATA_SMS"
	MessageTypeMMS           MessageType = "MMS"
	MessageTypeMmsDownloaded MessageType = "MMS_DOWNLOADED"
)

type MessageBody struct {
	ExtID       string
	Type        MessageType
	Sender      string
	Recipient   *string
	SimNumber   *uint8
	Content     string
	IsEncrypted bool
	CreatedAt   time.Time
}

// MessageInput holds the data needed to create an inbox message.
type MessageInput struct {
	MessageBody

	Attachments []AttachmentInput
}

// AttachmentInput holds the data needed to create an inbox attachment.
type AttachmentInput struct {
	PartID      int64
	ContentType string
	Name        string
	Size        *int64
	Data        []byte
}

// Message represents a stored inbox message.
type Message struct {
	MessageBody

	ID          uint64
	DeviceID    string
	Attachments []Attachment
}

// Attachment represents a stored inbox attachment.
type Attachment struct {
	AttachmentInput

	ID uint64
}

// ListFilter defines filters for listing inbox messages.
type ListFilter struct {
	DeviceID  string
	Type      MessageType
	StartDate time.Time
	EndDate   time.Time
}

func (f ListFilter) apply(query *gorm.DB) *gorm.DB {
	if f.DeviceID != "" {
		query = query.Where("device_id = ?", f.DeviceID)
	}
	if f.Type != "" {
		query = query.Where("type = ?", f.Type)
	}
	if !f.StartDate.IsZero() {
		query = query.Where("created_at >= ?", f.StartDate)
	}
	if !f.EndDate.IsZero() {
		query = query.Where("created_at <= ?", f.EndDate)
	}
	return query
}

// ListOptions defines pagination options for listing inbox messages.
type ListOptions struct {
	Limit  int
	Offset int
}

func (o ListOptions) apply(query *gorm.DB) *gorm.DB {
	if o.Limit > 0 {
		query = query.Limit(o.Limit)
	}
	if o.Offset > 0 {
		query = query.Offset(o.Offset)
	}
	return query
}
