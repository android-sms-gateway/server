package auth

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}

// GetByID returns a user by their ID.
func (r *repository) GetByID(id string) (models.User, error) {
	user := models.User{}

	return user, r.db.Where("id = ?", id).Take(&user).Error
}

func (r *repository) GetByLogin(login string) (models.User, error) {
	user := models.User{}

	return user, r.db.Where("id = ?", login).Take(&user).Error
}

func (r *repository) Insert(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *repository) UpdatePassword(userID string, passwordHash string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}
