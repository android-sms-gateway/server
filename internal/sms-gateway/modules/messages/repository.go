package messages

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const hashingLockName = "36444143-1ace-4dbf-891c-cc505911497e"

var ErrMessageNotFound = gorm.ErrRecordNotFound
var ErrMessageAlreadyExists = errors.New("duplicate id")

type repository struct {
	db *gorm.DB
}

func (r *repository) SelectPending(deviceID string) (messages []models.Message, err error) {
	err = r.db.
		Where("device_id = ? AND state = ?", deviceID, models.ProcessingStatePending).
		Order("priority DESC, id DESC").
		Limit(100).
		Preload("Recipients").
		Find(&messages).
		Error

	return
}

func (r *repository) Get(ID string, filter MessagesSelectFilter, options ...MessagesSelectOptions) (message models.Message, err error) {
	query := r.db.Model(&message).
		Where("ext_id = ?", ID)

	if filter.DeviceID != "" {
		query = query.Where("device_id = ?", filter.DeviceID)
	}

	if len(options) > 0 {
		if options[0].WithRecipients {
			query = query.Preload("Recipients")
		}
		if options[0].WithDevice {
			query = query.Joins("Device")
		}
		if options[0].WithStates {
			query = query.Preload("States")
		}
	}

	err = query.Take(&message).Error

	return
}

func (r *repository) Insert(message *models.Message) error {
	err := r.db.Omit("Device").Create(message).Error
	if err == nil {
		return nil
	}

	if mysqlErr := err.(*mysql.MySQLError); mysqlErr != nil && mysqlErr.Number == 1062 {
		return ErrMessageAlreadyExists
	}
	return err
}

func (r *repository) UpdateState(message *models.Message) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(message).Select("State").Updates(message).Error; err != nil {
			return err
		}

		for _, v := range message.States {
			v.MessageID = message.ID
			if err := tx.Model(&v).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&v).Error; err != nil {
				return err
			}
		}

		for _, v := range message.Recipients {
			if err := tx.Model(&v).Where("message_id = ?", message.ID).Select("State", "Error").Updates(&v).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repository) HashProcessed(ids []uint64) error {
	rawSQL := "UPDATE `messages` `m`, `message_recipients` `r`\n" +
		"SET `m`.`is_hashed` = true, `m`.`message` = SHA2(m.message, 256), `r`.`phone_number` = LEFT(SHA2(phone_number, 256), 16)\n" +
		"WHERE `m`.`id` = `r`.`message_id` AND `m`.`is_hashed` = false AND `m`.`is_encrypted` = false AND `m`.`state` <> 'Pending'"
	params := []interface{}{}
	if len(ids) > 0 {
		rawSQL += " AND `m`.`id` IN (?)"
		params = append(params, ids)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		hasLock := sql.NullBool{}
		lockRow := tx.Raw("SELECT GET_LOCK(?, 1)", hashingLockName).Row()
		err := lockRow.Scan(&hasLock)
		if err != nil {
			return err
		}

		if !hasLock.Valid || !hasLock.Bool {
			return errors.New("failed to acquire lock")
		}
		defer tx.Exec("SELECT RELEASE_LOCK(?)", hashingLockName)

		return tx.Exec(rawSQL, params...).Error
	})
}

// removeProcessed removes messages older than the given time that are not in
// the Pending state.
//
// This is useful for periodically cleaning up old messages that are not in the
// Pending state.
func (r *repository) removeProcessed(ctx context.Context, until time.Time) (int64, error) {
	res := r.db.
		WithContext(ctx).
		Where("state <> ?", models.ProcessingStatePending).
		Where("created_at < ?", until).
		Delete(&models.Message{})
	return res.RowsAffected, res.Error
}

func newRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}
