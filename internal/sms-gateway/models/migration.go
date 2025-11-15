package models

import (
	"embed"
	"fmt"

	"gorm.io/gorm"
)

//go:embed migrations
var migrations embed.FS

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(User), new(Device)); err != nil {
		return fmt.Errorf("models migration failed: %w", err)
	}
	return nil
}
