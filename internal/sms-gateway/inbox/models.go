package inbox

import (
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
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

func (*messageModel) TableName() string {
	return "inbox"
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

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(messageModel), new(attachmentModel)); err != nil {
		return fmt.Errorf("inbox migration failed: %w", err)
	}
	return nil
}
