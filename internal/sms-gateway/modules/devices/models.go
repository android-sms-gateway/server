package devices

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type DeviceSettings struct {
	UserID   string   `gorm:"primaryKey;not null;type:varchar(32)"`
	Settings Settings `gorm:"not null;type:json"`

	User models.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&DeviceSettings{})
}
