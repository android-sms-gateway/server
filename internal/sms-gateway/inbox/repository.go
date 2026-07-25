package inbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository provides data access for inbox messages.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Insert stores a single inbox message with its attachments in a transaction.
func (r *Repository) Insert(ctx context.Context, msg *messageModel) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Device", "Attachments").Create(msg).Error; err != nil {
			return fmt.Errorf("failed to insert inbox message: %w", err)
		}
		if len(msg.Attachments) > 0 {
			for i := range msg.Attachments {
				msg.Attachments[i].MessageID = msg.ID
			}
			if err := tx.Create(&msg.Attachments).Error; err != nil {
				return fmt.Errorf("failed to insert attachments: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to insert inbox message: %w", err)
	}
	return nil
}

// InsertBatch stores multiple inbox messages in a single transaction.
// Uses OnConflict{DoNothing: true} for idempotency on the (ext_id, device_id) unique index.
func (r *Repository) InsertBatch(ctx context.Context, msgs []*messageModel) error {
	if len(msgs) == 0 {
		return nil
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, msg := range msgs {
			if err := insertMessageWithAttachments(tx, msg); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to insert inbox message batch: %w", err)
	}
	return nil
}

func insertMessageWithAttachments(tx *gorm.DB, msg *messageModel) error {
	result := tx.Omit("Device", "Attachments").Clauses(
		clause.OnConflict{DoNothing: true},
	).Create(msg)
	if result.Error != nil {
		return fmt.Errorf("failed to insert inbox message: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil
	}
	if len(msg.Attachments) > 0 {
		for i := range msg.Attachments {
			msg.Attachments[i].MessageID = msg.ID
		}
		if err := tx.Create(&msg.Attachments).Error; err != nil {
			return fmt.Errorf("failed to insert attachments: %w", err)
		}
	}
	return nil
}

// list returns inbox messages for a user with filtering and pagination.
func (r *Repository) list(filter ListFilter, opts ListOptions) ([]messageModel, int64, error) {
	query := r.db.Model((*messageModel)(nil)).
		Joins("JOIN devices ON inbox.device_id = devices.id").
		Where("devices.user_id = ?", filter.UserID)

	if filter.DeviceID != "" {
		query = query.Where("inbox.device_id = ?", filter.DeviceID)
	}
	if filter.Type != "" {
		query = query.Where("inbox.type = ?", filter.Type)
	}
	if !filter.StartDate.IsZero() {
		query = query.Where("inbox.created_at >= ?", filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		query = query.Where("inbox.created_at < ?", filter.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count inbox messages: %w", err)
	}

	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	query = query.Order("inbox.created_at DESC, inbox.id DESC")

	var messages []messageModel
	if err := query.Preload("Attachments").Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list inbox messages: %w", err)
	}

	return messages, total, nil
}

// findMessageByExtID finds a message by external ID, verifying user ownership via device join.
func (r *Repository) findMessageByExtID(
	ctx context.Context,
	extID string,
	userID string,
) (*messageModel, error) {
	var msg messageModel
	err := r.db.WithContext(ctx).
		Joins("JOIN devices ON inbox.device_id = devices.id").
		Where("inbox.ext_id = ? AND devices.user_id = ?", extID, userID).
		Preload("Attachments").
		First(&msg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find inbox message: %w", err)
	}
	return &msg, nil
}

// findAttachment returns a single attachment by message ID and part ID.
func (r *Repository) findAttachment(
	ctx context.Context,
	messageID uint64,
	partID int64,
) (*attachmentModel, error) {
	var att attachmentModel
	err := r.db.WithContext(ctx).
		Where("message_id = ? AND part_id = ?", messageID, partID).
		First(&att).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find attachment: %w", err)
	}
	return &att, nil
}

// Cleanup deletes inbox messages older than the given time.
// Cascades to inbox_attachments via foreign key.
func (r *Repository) Cleanup(ctx context.Context, until time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("created_at < ?", until).
		Delete(new(messageModel))
	return res.RowsAffected, res.Error
}
