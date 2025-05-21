package webhooks

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type Webhook struct {
	ID     uint64 `json:"-"    gorm:"->;primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	ExtID  string `json:"id"   gorm:"not null;type:varchar(36);uniqueIndex:unq_webhooks_user_extid,priority:2"`
	UserID string `json:"-"    gorm:"<-:create;not null;type:varchar(32);uniqueIndex:unq_webhooks_user_extid,priority:1"`

	DeviceID *string `json:"device_id,omitempty" gorm:"type:varchar(21);index:idx_webhooks_device"`

	URL   string                  `json:"url"   validate:"required,http_url"   gorm:"not null;type:varchar(256)"`
	Event smsgateway.WebhookEvent `json:"event" gorm:"not null;type:varchar(32)"`

	User   models.User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Device *models.Device `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE"`

	models.SoftDeletableModel
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Webhook{})
}
