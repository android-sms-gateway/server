package settings

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repository struct {
	db *gorm.DB
}

// GetSettings retrieves the device settings for a user by their userID.
func (r *repository) GetSettings(userID string) (*DeviceSettings, error) {
	settings := &DeviceSettings{}
	err := r.db.Where("user_id = ?", userID).Limit(1).Find(settings).Error
	if err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateSettings updates the settings for a user.
func (r *repository) UpdateSettings(settings *DeviceSettings) (*DeviceSettings, error) {
	var updatedSettings *DeviceSettings
	err := r.db.Transaction(func(tx *gorm.DB) error {
		source := &DeviceSettings{UserID: settings.UserID}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(source).Error; err != nil {
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

		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(settings).Error; err != nil {
			return err
		}

		// Return the updated settings
		updatedSettings = settings
		return nil
	})
	return updatedSettings, err
}

// ReplaceSettings replaces the settings for a user.
//
// This function will overwrite all existing settings for the user.
func (r *repository) ReplaceSettings(settings *DeviceSettings) (*DeviceSettings, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Save(settings).Error
	})
	return settings, err
}

func newRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}
