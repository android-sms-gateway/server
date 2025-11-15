package settings

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repository struct {
	db *gorm.DB
}

// GetSettings retrieves the device settings for a user by their userID.
func (r *repository) GetSettings(userID string) (*DeviceSettings, error) {
	settings := new(DeviceSettings)
	err := r.db.Where("user_id = ?", userID).Limit(1).Find(settings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	if settings.Settings == nil {
		settings.Settings = map[string]any{}
	}

	return settings, nil
}

// UpdateSettings updates the settings for a user.
func (r *repository) UpdateSettings(settings *DeviceSettings) (*DeviceSettings, error) {
	var updatedSettings *DeviceSettings
	err := r.db.Transaction(func(tx *gorm.DB) error {
		source := new(DeviceSettings)
		if err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).Where("user_id = ?", settings.UserID).Limit(1).Find(source).Error; err != nil {
			return err
		}

		if source.Settings == nil {
			source.Settings = map[string]any{}
		}

		var err error
		settings.Settings, err = appendMap(source.Settings, settings.Settings, rules)
		if err != nil {
			return err
		}

		err = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(settings).Error
		if err != nil {
			return err
		}

		// Return the updated settings
		updatedSettings = settings
		return nil
	})
	if err != nil {
		return updatedSettings, fmt.Errorf("failed to update settings: %w", err)
	}

	return updatedSettings, nil
}

// ReplaceSettings replaces the settings for a user.
//
// This function will overwrite all existing settings for the user.
func (r *repository) ReplaceSettings(settings *DeviceSettings) (*DeviceSettings, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Save(settings).Error
	})

	if err != nil {
		return settings, fmt.Errorf("failed to replace settings: %w", err)
	}

	return settings, nil
}

func newRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}
