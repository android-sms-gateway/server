package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Insert(ctx context.Context, token *tokenModel) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("can't create token: %w", err)
	}

	return nil
}

func (r *Repository) Revoke(ctx context.Context, jti, userID string) error {
	if err := r.db.WithContext(ctx).Model((*tokenModel)(nil)).
		Where("id = ? and user_id = ? and revoked_at is null", jti, userID).
		Update("revoked_at", gorm.Expr("NOW()")).Error; err != nil {
		return fmt.Errorf("can't revoke token: %w", err)
	}

	return nil
}

func (r *Repository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model((*tokenModel)(nil)).
		Where("id = ? and revoked_at is not null", jti).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("can't check if token is revoked: %w", err)
	}

	return count > 0, nil
}

func (r *Repository) RotateRefreshToken(
	ctx context.Context,
	currentJTI string,
	nextRefresh, nextAccess *tokenModel,
) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current tokenModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", currentJTI).
			First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidToken
			}
			return fmt.Errorf("can't lock refresh token: %w", err)
		}

		now := time.Now()
		if current.RevokedAt != nil {
			return ErrTokenReplay
		}

		if current.ExpiresAt.Before(now) {
			return ErrInvalidToken
		}

		if err := tx.Model((*tokenModel)(nil)).Where("id = ?", current.ID).Updates(map[string]any{
			"revoked_at": now,
		}).Error; err != nil {
			return fmt.Errorf("can't mark refresh token as replaced: %w", err)
		}

		if err := tx.Create(nextRefresh).Error; err != nil {
			return fmt.Errorf("can't create next refresh token: %w", err)
		}

		if err := tx.Create(nextAccess).Error; err != nil {
			return fmt.Errorf("can't create next access token: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("can't rotate refresh token: %w", err)
	}

	return nil
}
