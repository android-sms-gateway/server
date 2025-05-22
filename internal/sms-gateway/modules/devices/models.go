package devices

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type DeviceSettings struct {
	UserID   string         `gorm:"primaryKey;not null;type:varchar(32)"`
	Settings map[string]any `gorm:"not null;type:json;serializer:json"`

	User models.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// Migrate performs automatic schema migration for the DeviceSettings table using GORM.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&DeviceSettings{})
}
