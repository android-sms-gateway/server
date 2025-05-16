package devices

import (
	"context"
	"errors"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = gorm.ErrRecordNotFound
	ErrInvalidFilter = errors.New("invalid filter")
	ErrMoreThanOne   = errors.New("more than one record")
)

type repository struct {
	db *gorm.DB
}

func (r *repository) Select(filter ...SelectFilter) ([]models.Device, error) {
	if len(filter) == 0 {
		return nil, ErrInvalidFilter
	}

	f := newFilter(filter...)
	devices := []models.Device{}

	return devices, f.apply(r.db).Find(&devices).Error
}

// Exists checks if there exists a device with the given filters.
//
// If the device does not exist, it returns false and nil error. If there is an
// error during the query, it returns false and the error. Otherwise, it returns
// true and nil error.
func (r *repository) Exists(filters ...SelectFilter) (bool, error) {
	err := newFilter(filters...).apply(r.db).Take(&models.Device{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *repository) Get(filter ...SelectFilter) (models.Device, error) {
	devices, err := r.Select(filter...)
	if err != nil {
		return models.Device{}, err
	}

	if len(devices) == 0 {
		return models.Device{}, ErrNotFound
	}

	if len(devices) > 1 {
		return models.Device{}, ErrMoreThanOne
	}

	return devices[0], nil
}

func (r *repository) Insert(device *models.Device) error {
	return r.db.Create(device).Error
}

func (r *repository) UpdatePushToken(id, token string) error {
	return r.db.Model(&models.Device{}).Where("id", id).Update("push_token", token).Error
}

func (r *repository) UpdateLastSeen(id string) error {
	return r.db.Model(&models.Device{}).Where("id", id).Update("last_seen", time.Now()).Error
}

func (r *repository) Remove(filter ...SelectFilter) error {
	if len(filter) == 0 {
		return ErrInvalidFilter
	}

	f := newFilter(filter...)
	return f.apply(r.db).Delete(&models.Device{}).Error
}

func (r *repository) removeUnused(ctx context.Context, since time.Time) (int64, error) {
	res := r.db.
		WithContext(ctx).
		Where("updated_at < ?", since).
		Delete(&models.Device{})

	return res.RowsAffected, res.Error
}

func newDevicesRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}
