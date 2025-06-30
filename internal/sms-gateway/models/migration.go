package models

import (
	"embed"

	"gorm.io/gorm"
)

//go:embed migrations
var migrations embed.FS

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Device{})
}
