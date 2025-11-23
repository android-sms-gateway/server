package users

import (
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type userModel struct {
	models.SoftDeletableModel

	ID           string `gorm:"primaryKey;type:varchar(32)"`
	PasswordHash string `gorm:"not null;type:varchar(72)"`
}

func newUserModel(id string, passwordHash string) *userModel {
	//nolint:exhaustruct // partial constructor
	return &userModel{
		ID:           id,
		PasswordHash: passwordHash,
	}
}

func (u *userModel) TableName() string {
	return "users"
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(userModel)); err != nil {
		return fmt.Errorf("users migration failed: %w", err)
	}
	return nil
}
