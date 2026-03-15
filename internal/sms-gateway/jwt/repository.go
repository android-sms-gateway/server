package jwt

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
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

func (r *Repository) RevokeByUser(ctx context.Context, userID string) (int64, error) {
	var res *gorm.DB

	if res = r.db.WithContext(ctx).Model((*tokenModel)(nil)).
		Where("user_id = ? and revoked_at is null and expires_at > NOW()", userID).
		Update("revoked_at", gorm.Expr("NOW()")); res.Error != nil {
		return 0, fmt.Errorf("can't revoke user tokens: %w", res.Error)
	}

	return res.RowsAffected, nil
}
