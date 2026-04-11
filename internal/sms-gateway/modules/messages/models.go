package messages

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/samber/lo"
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

type messageModel struct {
	models.SoftDeletableModel

	ID                 uint64          `gorm:"primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	DeviceID           string          `gorm:"not null;type:char(21);uniqueIndex:unq_messages_id_device,priority:2;index:idx_messages_device_state"`
	ExtID              string          `gorm:"not null;type:varchar(36);uniqueIndex:unq_messages_id_device,priority:1"`
	Type               MessageType     `gorm:"not null;type:enum('Text','Data');default:Text"`
	Content            string          `gorm:"not null;type:text"`
	State              ProcessingState `gorm:"not null;type:enum('Pending','Sent','Processed','Delivered','Failed');default:Pending;index:idx_messages_device_state"`
	ValidUntil         *time.Time      `gorm:"type:datetime"`
	ScheduleAt         *time.Time      `gorm:"type:datetime"`
	SimNumber          *uint8          `gorm:"type:tinyint(1) unsigned"`
	WithDeliveryReport bool            `gorm:"not null;type:tinyint(1) unsigned"`
	Priority           int8            `gorm:"not null;type:tinyint;default:0"`

	IsHashed    bool `gorm:"not null;type:tinyint(1) unsigned;default:0"`
	IsEncrypted bool `gorm:"not null;type:tinyint(1) unsigned;default:0"`

	Device     models.Device           `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE"`
	Recipients []messageRecipientModel `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
	States     []messageStateModel     `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
}

func newMessageModel(
	extID string,
	deviceID string,
	phoneNumbers []string,
	priority int8,
	simNumber *uint8,
	validUntil *time.Time,
	scheduleAt *time.Time,
	withDeliveryReport bool,
	isEncrypted bool,
) *messageModel {
	//nolint:exhaustruct // partial constructor
	return &messageModel{
		ExtID:    extID,
		DeviceID: deviceID,
		Recipients: lo.Map(phoneNumbers, func(item string, _ int) messageRecipientModel {
			return newMessageRecipient(item, ProcessingStatePending, nil)
		}),
		Priority:           priority,
		SimNumber:          simNumber,
		ValidUntil:         validUntil,
		ScheduleAt:         scheduleAt,
		WithDeliveryReport: withDeliveryReport,
		IsEncrypted:        isEncrypted,

		State: ProcessingStatePending,
	}
}

func (*messageModel) TableName() string {
	return "messages"
}

func (m *messageModel) SetTextContent(content TextMessageContent) error {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	m.Type = MessageTypeText
	m.Content = string(contentJSON)

	return nil
}

func (m *messageModel) GetTextContent() (*TextMessageContent, error) {
	if m.Type != MessageTypeText || m.Content == "" || m.IsHashed {
		return nil, nil //nolint:nilnil // special meaning
	}

	content := new(TextMessageContent)

	err := json.Unmarshal([]byte(m.Content), content)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal text content: %w", err)
	}

	return content, nil
}

func (m *messageModel) SetDataContent(content DataMessageContent) error {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	m.Type = MessageTypeData
	m.Content = string(contentJSON)

	return nil
}

func (m *messageModel) GetDataContent() (*DataMessageContent, error) {
	if m.Type != MessageTypeData || m.Content == "" || m.IsHashed {
		return nil, nil //nolint:nilnil // special meaning
	}

	content := new(DataMessageContent)

	err := json.Unmarshal([]byte(m.Content), content)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data content: %w", err)
	}

	return content, nil
}

func (m *messageModel) GetHashedContent() (*HashedMessageContent, error) {
	if !m.IsHashed || m.Content == "" {
		return nil, nil //nolint:nilnil // special meaning
	}

	return &smsgateway.HashedMessage{
		Hash: m.Content,
	}, nil
}

func (m *messageModel) toStateDomain() (*MessageState, error) {
	textContent, err := m.GetTextContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode text content: %w", err)
	}

	dataContent, err := m.GetDataContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode data content: %w", err)
	}

	hashedContent, err := m.GetHashedContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode hashed content: %w", err)
	}

	content := MessageStateContent{
		MessageContent: MessageContent{
			TextContent: textContent,
			DataContent: dataContent,
		},
		HashedContent: hashedContent,
	}

	return &MessageState{
		MessageStateInput: MessageStateInput{
			ID:    m.ExtID,
			State: m.State,
			Recipients: lo.Map(
				m.Recipients,
				func(item messageRecipientModel, _ int) smsgateway.RecipientState {
					return item.toDomain()
				},
			),
			States: lo.Associate(
				m.States,
				func(item messageStateModel) (string, time.Time) { return string(item.State), item.UpdatedAt },
			),
		},
		MessageStateContent: content,

		DeviceID:    m.DeviceID,
		IsHashed:    m.IsHashed,
		IsEncrypted: m.IsEncrypted,
	}, nil
}

type messageRecipientModel struct {
	ID          uint64          `gorm:"primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	MessageID   uint64          `gorm:"uniqueIndex:unq_message_recipients_message_id_phone_number,priority:1;type:BIGINT UNSIGNED"`
	PhoneNumber string          `gorm:"uniqueIndex:unq_message_recipients_message_id_phone_number,priority:2;type:varchar(128)"`
	State       ProcessingState `gorm:"not null;type:enum('Pending','Sent','Processed','Delivered','Failed');default:Pending"`
	Error       *string         `gorm:"type:varchar(256)"`
}

func newMessageRecipient(phoneNumber string, state ProcessingState, err *string) messageRecipientModel {
	return messageRecipientModel{
		ID:          0,
		MessageID:   0,
		PhoneNumber: phoneNumber,
		State:       state,
		Error:       err,
	}
}

func (m *messageRecipientModel) TableName() string {
	return "message_recipients"
}

func (m *messageRecipientModel) toDomain() smsgateway.RecipientState {
	return smsgateway.RecipientState{
		PhoneNumber: m.PhoneNumber,
		State:       smsgateway.ProcessingState(m.State),
		Error:       m.Error,
	}
}

type messageStateModel struct {
	ID        uint64          `gorm:"primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	MessageID uint64          `gorm:"not null;type:BIGINT UNSIGNED;uniqueIndex:unq_message_states_message_id_state,priority:1"`
	State     ProcessingState `gorm:"not null;type:enum('Pending','Sent','Processed','Delivered','Failed');uniqueIndex:unq_message_states_message_id_state,priority:2"`
	UpdatedAt time.Time       `gorm:"<-:create;not null;autoupdatetime:false"`
}

func (m *messageStateModel) TableName() string {
	return "message_states"
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(messageModel), new(messageRecipientModel), new(messageStateModel)); err != nil {
		return fmt.Errorf("messages migration failed: %w", err)
	}
	return nil
}
