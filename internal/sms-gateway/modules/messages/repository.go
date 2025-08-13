package messages

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const hashingLockName = "36444143-1ace-4dbf-891c-cc505911497e"
const maxPendingBatch = 100

var ErrMessageNotFound = gorm.ErrRecordNotFound
var ErrMessageAlreadyExists = errors.New("duplicate id")
var ErrMultipleMessagesFound = errors.New("multiple messages found")

type repository struct {
	db *gorm.DB
}

func (r *repository) Select(filter MessagesSelectFilter, options MessagesSelectOptions) ([]Message, int64, error) {
	query := r.db.Model(&Message{})

	// Apply date range filter
	if !filter.StartDate.IsZero() {
		query = query.Where("messages.created_at >= ?", filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		query = query.Where("messages.created_at < ?", filter.EndDate)
	}

	// Apply ID filter
	if filter.ExtID != "" {
		query = query.Where("messages.ext_id = ?", filter.ExtID)
	}

	// Apply user filter
	if filter.UserID != "" {
		query = query.
			Joins("JOIN devices ON messages.device_id = devices.id").
			Where("devices.user_id = ?", filter.UserID)
	}

	// Apply state filter
	if filter.State != "" {
		query = query.Where("messages.state = ?", filter.State)
	}

	// Apply device filter
	if filter.DeviceID != "" {
		query = query.Where("messages.device_id = ?", filter.DeviceID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset > 0 {
		query = query.Offset(options.Offset)
	}

	// Apply ordering
	if options.OrderBy == MessagesOrderFIFO {
		query = query.Order("messages.priority DESC, messages.id ASC")
	} else {
		query = query.Order("messages.priority DESC, messages.id DESC")
	}

	// Preload related data
	if options.WithRecipients {
		query = query.Preload("Recipients")
	}
	if filter.UserID == "" && options.WithDevice {
		query = query.Joins("Device")
	}
	if options.WithStates {
		query = query.Preload("States")
	}

	messages := make([]Message, 0, min(options.Limit, int(total)))
	if err := query.Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("can't select messages: %w", err)
	}

	return messages, total, nil
}

func (r *repository) SelectPending(deviceID string, order MessagesOrder) ([]Message, error) {
	messages, _, err := r.Select(MessagesSelectFilter{
		DeviceID: deviceID,
		State:    ProcessingStatePending,
	}, MessagesSelectOptions{
		WithRecipients: true,
		Limit:          maxPendingBatch,
		OrderBy:        order,
	})

	return messages, err
}

func (r *repository) Get(filter MessagesSelectFilter, options MessagesSelectOptions) (Message, error) {
	messages, _, err := r.Select(filter, options)
	if err != nil {
		return Message{}, fmt.Errorf("can't get message: %w", err)
	}

	if len(messages) == 0 {
		return Message{}, ErrMessageNotFound
	}

	if len(messages) > 1 {
		return Message{}, ErrMultipleMessagesFound
	}

	return messages[0], nil
}

func (r *repository) Insert(message *Message) error {
	err := r.db.Omit("Device").Create(message).Error
	if err == nil {
		return nil
	}

	if mysqlErr := err.(*mysql.MySQLError); mysqlErr != nil && mysqlErr.Number == 1062 {
		return ErrMessageAlreadyExists
	}
	return err
}

func (r *repository) UpdateState(message *Message) error {
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
		"SET `m`.`is_hashed` = true, `m`.`content` = SHA2(COALESCE(JSON_VALUE(`content`, '$.text'), JSON_VALUE(`content`, '$.data')), 256), `r`.`phone_number` = LEFT(SHA2(phone_number, 256), 16)\n" +
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
		Where("state <> ?", ProcessingStatePending).
		Where("created_at < ?", until).
		Delete(&Message{})
	return res.RowsAffected, res.Error
}

func newRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}
