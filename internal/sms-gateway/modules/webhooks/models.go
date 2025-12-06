package webhooks

import (
	"fmt"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/users"
	"gorm.io/gorm"
)

type Webhook struct {
	models.SoftDeletableModel

	ID     uint64 `json:"-"  gorm:"->;primaryKey;type:BIGINT UNSIGNED;autoIncrement"`
	ExtID  string `json:"id" gorm:"not null;type:varchar(36);uniqueIndex:unq_webhooks_user_extid,priority:2"`
	UserID string `json:"-"  gorm:"<-:create;not null;type:varchar(32);uniqueIndex:unq_webhooks_user_extid,priority:1"`

	DeviceID *string `json:"device_id,omitempty" gorm:"type:varchar(21);index:idx_webhooks_device"`

	URL   string                  `json:"url"   validate:"required,http_url" gorm:"not null;type:varchar(256)"`
	Event smsgateway.WebhookEvent `json:"event"                              gorm:"not null;type:varchar(32)"`

	User   users.User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Device *models.Device `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE"`
}

func newWebhook(extID string, url string, event smsgateway.WebhookEvent, userID string, deviceID *string) *Webhook {
	//nolint:exhaustruct // partial constructor
	return &Webhook{
		ExtID:    extID,
		URL:      url,
		Event:    event,
		UserID:   userID,
		DeviceID: deviceID,
	}
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(Webhook)); err != nil {
		return fmt.Errorf("webhooks migration failed: %w", err)
	}
	return nil
}
