package jwt

import (
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type tokenUse string

const (
	accessToken  tokenUse = "access"
	refreshToken tokenUse = "refresh"
)

type tokenModel struct {
	models.TimedModel

	ID        string    `gorm:"primaryKey;type:char(21)"`
	UserID    string    `gorm:"not null;type:char(21);index:idx_tokens_user_id"`
	TokenUse  tokenUse  `gorm:"not null;type:enum('access','refresh');default:access"`
	ParentJTI *string   `gorm:"type:char(21);index:idx_tokens_parent_jti"`
	ExpiresAt time.Time `gorm:"not null;index:idx_tokens_expires_at"`
	RevokedAt *time.Time
}

func newAccessTokenModel(userID string, token TokenInfo) *tokenModel {
	//nolint:exhaustruct // partial constructor
	return &tokenModel{
		ID:        token.ID,
		UserID:    userID,
		TokenUse:  accessToken,
		ExpiresAt: token.ExpiresAt,
	}
}

func newRefreshTokenModel(userID string, parentJTI string, token TokenInfo) *tokenModel {
	//nolint:exhaustruct // partial constructor
	return &tokenModel{
		ID:        token.ID,
		UserID:    userID,
		TokenUse:  refreshToken,
		ParentJTI: &parentJTI,
		ExpiresAt: token.ExpiresAt,
	}
}

func (tokenModel) TableName() string {
	return "tokens"
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(tokenModel)); err != nil {
		return fmt.Errorf("tokens migration failed: %w", err)
	}
	return nil
}
