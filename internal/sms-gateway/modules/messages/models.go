package messages

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type ProcessingState string
type MessageType string

const (
	ProcessingStatePending   ProcessingState = "Pending"
	ProcessingStateProcessed ProcessingState = "Processed"
	ProcessingStateSent      ProcessingState = "Sent"
	ProcessingStateDelivered ProcessingState = "Delivered"
	ProcessingStateFailed    ProcessingState = "Failed"

	MessageTypeText MessageType = "Text"
	MessageTypeData MessageType = "Data"
)

type TextMessageContent struct {
	Text string `json:"text"`
}

type DataMessageContent struct {
	Data string `json:"data"`
	Port uint16 `json:"port"`
}

type Message struct {
	ID                 uint64          `gorm:"primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	DeviceID           string          `gorm:"not null;type:char(21);uniqueIndex:unq_messages_id_device,priority:2;index:idx_messages_device_state"`
	ExtID              string          `gorm:"not null;type:varchar(36);uniqueIndex:unq_messages_id_device,priority:1"`
	Type               MessageType     `gorm:"not null;type:enum('Text','Data');default:Text"`
	Content            string          `gorm:"not null;type:text"`
	State              ProcessingState `gorm:"not null;type:enum('Pending','Sent','Processed','Delivered','Failed');default:Pending;index:idx_messages_device_state"`
	ValidUntil         *time.Time      `gorm:"type:datetime"`
	SimNumber          *uint8          `gorm:"type:tinyint(1) unsigned"`
	WithDeliveryReport bool            `gorm:"not null;type:tinyint(1) unsigned"`
	Priority           int8            `gorm:"not null;type:tinyint;default:0"`

	IsHashed    bool `gorm:"not null;type:tinyint(1) unsigned;default:0"`
	IsEncrypted bool `gorm:"not null;type:tinyint(1) unsigned;default:0"`

	Device     models.Device      `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE"`
	Recipients []MessageRecipient `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
	States     []MessageState     `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`

	models.SoftDeletableModel
}

func (m *Message) SetTextContent(content TextMessageContent) error {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return err
	}

	m.Type = MessageTypeText
	m.Content = string(contentJSON)

	return nil
}

func (m *Message) GetTextContent() (*TextMessageContent, error) {
	if m.Type != MessageTypeText {
		return nil, nil
	}

	content := TextMessageContent{}

	err := json.Unmarshal([]byte(m.Content), &content)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal text content: %w", err)
	}

	return &content, nil
}

func (m *Message) SetDataContent(content DataMessageContent) error {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return err
	}

	m.Type = MessageTypeData
	m.Content = string(contentJSON)

	return nil
}

func (m *Message) GetDataContent() (*DataMessageContent, error) {
	if m.Type != MessageTypeData {
		return nil, nil
	}

	content := DataMessageContent{}

	err := json.Unmarshal([]byte(m.Content), &content)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data content: %w", err)
	}

	return &content, nil
}

type MessageRecipient struct {
	ID          uint64          `gorm:"primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	MessageID   uint64          `gorm:"uniqueIndex:unq_message_recipients_message_id_phone_number,priority:1;type:BIGINT UNSIGNED"`
	PhoneNumber string          `gorm:"uniqueIndex:unq_message_recipients_message_id_phone_number,priority:2;type:varchar(128)"`
	State       ProcessingState `gorm:"not null;type:enum('Pending','Sent','Processed','Delivered','Failed');default:Pending"`
	Error       *string         `gorm:"type:varchar(256)"`
}

type MessageState struct {
	ID        uint64          `gorm:"primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	MessageID uint64          `gorm:"not null;type:BIGINT UNSIGNED;uniqueIndex:unq_message_states_message_id_state,priority:1"`
	State     ProcessingState `gorm:"not null;type:enum('Pending','Sent','Processed','Delivered','Failed');uniqueIndex:unq_message_states_message_id_state,priority:2"`
	UpdatedAt time.Time       `gorm:"<-:create;not null;autoupdatetime:false"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Message{}, &MessageRecipient{}, &MessageState{})
}
