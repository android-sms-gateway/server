package settings

import (
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type DeviceSettings struct {
	UserID   string         `gorm:"primaryKey;not null;type:varchar(32)"`
	Settings map[string]any `gorm:"not null;type:json;serializer:json"`

	User models.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	models.TimedModel
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&DeviceSettings{}); err != nil {
		return fmt.Errorf("device_settings migration failed: %w", err)
	}
	return nil
}
