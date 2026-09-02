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

// InsertBatch stores multiple inbox messages in a single transaction.
// Uses OnConflict{DoNothing: true} for idempotency on the (ext_id, device_id) unique index.
func (r *Repository) InsertBatch(ctx context.Context, deviceID string, msgs []MessageInput) error {
	if len(msgs) == 0 {
		return nil
	}

	models := make([]*messageModel, 0, len(msgs))
	for _, msg := range msgs {
		models = append(models, newMessageModel(deviceID, msg))
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, msg := range models {
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
func (r *Repository) list(userID string, filter ListFilter, opts ListOptions) ([]Message, int64, error) {
	query := r.db.Model((*messageModel)(nil)).
		Joins("JOIN devices ON inbox.device_id = devices.id").
		Where("devices.user_id = ?", userID)

	query = filter.apply(query)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count inbox messages: %w", err)
	}

	query = opts.apply(query)

	query = query.Order("inbox.created_at DESC, inbox.id DESC")

	var messages []messageModel
	if err := query.Preload("Attachments").Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list inbox messages: %w", err)
	}

	result := make([]Message, 0, len(messages))
	for i := range messages {
		result = append(result, messages[i].toDomain())
	}

	return result, total, nil
}

// findMessageByExtID finds a message by external ID, verifying user ownership via device join.
func (r *Repository) findMessageByExtID(
	ctx context.Context,
	extID string,
	userID string,
) (*Message, error) {
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
	domain := msg.toDomain()
	return &domain, nil
}

// findAttachment returns a single attachment by message ID and part ID.
func (r *Repository) findAttachment(
	ctx context.Context,
	messageID uint64,
	partID int64,
) (*Attachment, error) {
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
	domain := att.toDomain()
	return &domain, nil
}

// Cleanup deletes inbox messages older than the given time.
// Cascades to inbox_attachments via foreign key.
func (r *Repository) Cleanup(ctx context.Context, until time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("created_at < ?", until).
		Delete(new(messageModel))
	return res.RowsAffected, res.Error
}
