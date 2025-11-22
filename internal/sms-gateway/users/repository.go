package users

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

// newRepository creates a new repository instance.
func newRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Exists(id string) (bool, error) {
	var count int64
	if err := r.db.Model((*userModel)(nil)).
		Where("id = ?", id).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("can't check if user exists: %w", err)
	}

	return count > 0, nil
}

// GetByID retrieves a user by their ID.
func (r *repository) GetByID(id string) (*userModel, error) {
	user := new(userModel)

	if err := r.db.Where("id = ?", id).Take(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("can't get user: %w", err)
	}

	return user, nil
}

func (r *repository) Insert(user *userModel) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("can't create user: %w", err)
	}

	return nil
}

func (r *repository) UpdatePassword(ID string, passwordHash string) error {
	if err := r.db.Model((*userModel)(nil)).
		Where("id = ?", ID).
		Update("password_hash", passwordHash).Error; err != nil {
		return fmt.Errorf("can't update password: %w", err)
	}

	return nil
}
