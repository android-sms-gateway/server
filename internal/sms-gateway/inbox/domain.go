package inbox

import "time"

// MessageType represents the type of inbox message.
type MessageType string

const (
	MessageTypeSMS           MessageType = "SMS"
	MessageTypeDataSMS       MessageType = "DATA_SMS"
	MessageTypeMMS           MessageType = "MMS"
	MessageTypeMmsDownloaded MessageType = "MMS_DOWNLOADED"
)

// MessageInput holds the data needed to create an inbox message.
type MessageInput struct {
	ID          string
	Type        MessageType
	Sender      string
	Recipient   *string
	SimNumber   *uint8
	Content     string
	IsEncrypted bool
	CreatedAt   time.Time
	Attachments []AttachmentInput
}

// AttachmentInput holds the data needed to create an inbox attachment.
type AttachmentInput struct {
	PartID      int64
	ContentType string
	Name        string
	Size        *int64
	Data        []byte
	IsEncrypted bool
}

// Message represents a stored inbox message.
type Message struct {
	ID          uint64
	ExtID       string
	DeviceID    string
	Type        MessageType
	Sender      string
	Recipient   *string
	SimNumber   *uint8
	Content     string
	IsEncrypted bool
	CreatedAt   time.Time
	Attachments []Attachment
}

// Attachment represents a stored inbox attachment.
type Attachment struct {
	ID          uint64
	PartID      int64
	ContentType string
	Name        string
	Size        *int64
	Data        []byte
}

// ListFilter defines filters for listing inbox messages.
type ListFilter struct {
	UserID    string
	DeviceID  string
	Type      MessageType
	StartDate time.Time
	EndDate   time.Time
}

// ListOptions defines pagination options for listing inbox messages.
type ListOptions struct {
	Limit  int
	Offset int
}
