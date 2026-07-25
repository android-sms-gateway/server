package inbox

import (
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type messageModel struct {
	models.TimedModel

	ID          uint64      `gorm:"primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	ExtID       string      `gorm:"not null;type:varchar(36);uniqueIndex:unq_inbox_ext_device,priority:1"`
	DeviceID    string      `gorm:"not null;type:char(21);uniqueIndex:unq_inbox_ext_device,priority:2;index:idx_inbox_device_created"`
	Type        MessageType `gorm:"not null;type:enum('SMS','DATA_SMS','MMS','MMS_DOWNLOADED')"`
	Sender      string      `gorm:"not null;type:varchar(512)"`
	Recipient   *string     `gorm:"type:varchar(512)"`
	SimNumber   *uint8      `gorm:"type:tinyint(1) unsigned"`
	Content     string      `gorm:"not null;type:text"`
	IsEncrypted bool        `gorm:"not null;type:tinyint(1) unsigned;default:1"`

	Attachments []attachmentModel `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
}

func newMessageModel(deviceID string, input MessageInput) *messageModel {
	return &messageModel{
		TimedModel: models.TimedModel{
			CreatedAt: input.CreatedAt,
			UpdatedAt: input.CreatedAt,
		},

		ID:          0,
		ExtID:       input.ExtID,
		DeviceID:    deviceID,
		Type:        input.Type,
		Sender:      input.Sender,
		Recipient:   input.Recipient,
		SimNumber:   input.SimNumber,
		Content:     input.Content,
		IsEncrypted: input.IsEncrypted,

		Attachments: lo.Map(
			input.Attachments,
			func(item AttachmentInput, _ int) attachmentModel { return newAttachmentModel(item) },
		),
	}
}

func newAttachmentModel(input AttachmentInput) attachmentModel {
	return attachmentModel{
		ID:          0,
		MessageID:   0,
		PartID:      input.PartID,
		ContentType: input.ContentType,
		Name:        input.Name,
		Size:        input.Size,
		Data:        input.Data,
	}
}

func (*messageModel) TableName() string {
	return "inbox"
}

func (m *messageModel) toDomain() Message {
	return Message{
		MessageBody: MessageBody{
			ExtID:       m.ExtID,
			Type:        m.Type,
			Sender:      m.Sender,
			Recipient:   m.Recipient,
			SimNumber:   m.SimNumber,
			Content:     m.Content,
			IsEncrypted: m.IsEncrypted,
			CreatedAt:   m.CreatedAt,
		},

		ID:       m.ID,
		DeviceID: m.DeviceID,
		Attachments: lo.Map(
			m.Attachments,
			func(item attachmentModel, _ int) Attachment { return item.toDomain() },
		),
	}
}

type attachmentModel struct {
	ID          uint64 `gorm:"primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	MessageID   uint64 `gorm:"not null;type:BIGINT UNSIGNED;uniqueIndex:unq_inbox_att_msg_part,priority:1"`
	PartID      int64  `gorm:"not null;type:BIGINT;uniqueIndex:unq_inbox_att_msg_part,priority:2"`
	ContentType string `gorm:"not null;type:varchar(128)"`
	Name        string `gorm:"not null;type:varchar(512)"`
	Size        *int64 `gorm:"type:BIGINT"`
	Data        []byte `gorm:"not null;type:LONGBLOB"`
}

func (*attachmentModel) TableName() string {
	return "inbox_attachments"
}

func (m *attachmentModel) toDomain() Attachment {
	return Attachment{
		AttachmentInput: AttachmentInput{
			PartID:      m.PartID,
			ContentType: m.ContentType,
			Name:        m.Name,
			Size:        m.Size,
			Data:        m.Data,
		},

		ID: m.ID,
	}
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(messageModel), new(attachmentModel)); err != nil {
		return fmt.Errorf("inbox migration failed: %w", err)
	}
	return nil
}
