package jwt

import (
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type tokenModel struct {
	models.TimedModel

	ID        string    `gorm:"primaryKey;type:char(21)"`
	UserID    string    `gorm:"not null;type:char(21);index:idx_tokens_user_id"`
	ExpiresAt time.Time `gorm:"not null;index:idx_tokens_expires_at"`
	RevokedAt *time.Time
}

func (tokenModel) TableName() string {
	return "tokens"
}

func newTokenModel(id, userID string, expiresAt time.Time) *tokenModel {
	return &tokenModel{
		ID:        id,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(tokenModel)); err != nil {
		return fmt.Errorf("tokens migration failed: %w", err)
	}
	return nil
}
