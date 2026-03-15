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

func (r *Repository) Insert(ctx context.Context, tokens ...tokenModel) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.persistTokens(tx, tokens...); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("can't insert tokens: %w", err)
	}

	return nil
}

func (r *Repository) Revoke(ctx context.Context, jti, userID string) error {
	if err := r.db.WithContext(ctx).Model((*tokenModel)(nil)).
		Where("(id = ? or parent_jti = ?) and user_id = ? and revoked_at is null", jti, jti, userID).
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
	nextRefresh, nextAccess tokenModel,
) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current tokenModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND token_use = ?", currentJTI, refreshToken).
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
			"revoked_at": gorm.Expr("NOW()"),
		}).Error; err != nil {
			return fmt.Errorf("can't mark refresh token as replaced: %w", err)
		}

		if err := r.persistTokens(tx, nextRefresh, nextAccess); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("can't rotate refresh token: %w", err)
	}

	return nil
}

func (r *Repository) persistTokens(tx *gorm.DB, tokens ...tokenModel) error {
	if err := tx.Create(tokens).Error; err != nil {
		return fmt.Errorf("can't create tokens: %w", err)
	}

	return nil
}

func (r *Repository) RevokeByUser(ctx context.Context, userID string) (int64, error) {
	var res *gorm.DB

	if res = r.db.WithContext(ctx).Model((*tokenModel)(nil)).
		Where("user_id = ? and revoked_at is null and expires_at > NOW()", userID).
		Update("revoked_at", gorm.Expr("NOW()")); res.Error != nil {
		return 0, fmt.Errorf("can't revoke user tokens: %w", res.Error)
	}

	return res.RowsAffected, nil
}
