package models

import (
	"time"
)

type ProcessingState string

const (
	ProcessingStatePending   ProcessingState = "Pending"
	ProcessingStateProcessed ProcessingState = "Processed"
	ProcessingStateSent      ProcessingState = "Sent"
	ProcessingStateDelivered ProcessingState = "Delivered"
	ProcessingStateFailed    ProcessingState = "Failed"
)

type TimedModel struct {
	CreatedAt time.Time  `gorm:"->;not null;autocreatetime:false;default:CURRENT_TIMESTAMP(3)"`
	UpdatedAt time.Time  `gorm:"->;not null;autoupdatetime:false;default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
	DeletedAt *time.Time `gorm:"<-:update"`
}

type User struct {
	ID           string   `gorm:"primaryKey;type:varchar(32)"`
	PasswordHash string   `gorm:"not null;type:varchar(72)"`
	Devices      []Device `gorm:"-,foreignKey:UserID;constraint:OnDelete:CASCADE"`

	TimedModel
}

type Device struct {
	ID        string  `gorm:"primaryKey;type:char(21)"`
	Name      *string `gorm:"type:varchar(128)"`
	AuthToken string  `gorm:"not null;uniqueIndex;type:char(21)"`
	PushToken *string `gorm:"type:varchar(256)"`

	LastSeen time.Time `gorm:"not null;autocreatetime:false;default:CURRENT_TIMESTAMP(3)"`

	UserID string `gorm:"not null;type:varchar(32)"`

	TimedModel
}

func (d *Device) IsEmpty() bool {
	if d == nil {
		return true
	}

	return d.ID == ""
}

type Message struct {
	ID                 uint64          `gorm:"primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	DeviceID           string          `gorm:"not null;type:char(21);uniqueIndex:unq_messages_id_device,priority:2;index:idx_messages_device_state"`
	ExtID              string          `gorm:"not null;type:varchar(36);uniqueIndex:unq_messages_id_device,priority:1"`
	Message            string          `gorm:"not null;type:text"`
	State              ProcessingState `gorm:"not null;type:enum('Pending','Sent','Processed','Delivered','Failed');default:Pending;index:idx_messages_device_state"`
	ValidUntil         *time.Time      `gorm:"type:datetime"`
	SimNumber          *uint8          `gorm:"type:tinyint(1) unsigned"`
	WithDeliveryReport bool            `gorm:"not null;type:tinyint(1) unsigned"`
	Priority           int8            `gorm:"not null;type:tinyint;default:0"`

	IsHashed    bool `gorm:"not null;type:tinyint(1) unsigned;default:0"`
	IsEncrypted bool `gorm:"not null;type:tinyint(1) unsigned;default:0"`

	Device     Device             `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE"`
	Recipients []MessageRecipient `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
	States     []MessageState     `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`

	TimedModel
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
