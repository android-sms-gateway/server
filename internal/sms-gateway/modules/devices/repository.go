package devices

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrInvalidFilter = errors.New("invalid filter")
	ErrMoreThanOne   = errors.New("more than one record")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Select(ctx context.Context, filter ...SelectFilter) ([]Device, error) {
	if len(filter) == 0 {
		return nil, ErrInvalidFilter
	}

	f := newFilter(filter...)
	devices := []DeviceModel{}
	if err := f.apply(r.db.WithContext(ctx)).Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("failed to select devices: %w", err)
	}

	return lo.Map(devices, func(m DeviceModel, _ int) Device { return *m.toDomain() }), nil
}

// Exists checks if there exists a device with the given filters.
//
// If the device does not exist, it returns false and nil error. If there is an
// error during the query, it returns false and the error. Otherwise, it returns
// true and nil error.
func (r *Repository) Exists(ctx context.Context, filters ...SelectFilter) (bool, error) {
	if len(filters) == 0 {
		return false, ErrInvalidFilter
	}

	err := newFilter(filters...).apply(r.db.WithContext(ctx)).Take(new(DeviceModel)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *Repository) Get(ctx context.Context, filter ...SelectFilter) (*Device, error) {
	devices, err := r.Select(ctx, filter...)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	if len(devices) == 0 {
		return nil, ErrNotFound
	}

	if len(devices) > 1 {
		return nil, ErrMoreThanOne
	}

	return &devices[0], nil
}

func (r *Repository) Insert(ctx context.Context, device DeviceInput) (*Device, error) {
	model := newDeviceModel(device)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, fmt.Errorf("failed to insert device: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Repository) Update(ctx context.Context, id string, device DeviceUpdate) error {
	updates := map[string]any{}

	if device.PushToken != nil {
		updates["push_token"] = device.PushToken
	}

	if device.SimCards != nil {
		updates["sim_cards"] = datatypes.NewJSONSlice(lo.Map(
			device.SimCards,
			func(simCard SimCard, _ int) simCardModel { return newSimCardModel(simCard) },
		))
	}

	if len(updates) == 0 {
		return nil
	}

	err := r.db.
		WithContext(ctx).
		Model((*DeviceModel)(nil)).
		Where("id = ?", id).
		Updates(updates).
		Error
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	return nil
}

func (r *Repository) SetLastSeenBatch(ctx context.Context, batch map[string]time.Time) error {
	if len(batch) == 0 {
		return nil
	}
	const batchChunkSize = 500

	type entry struct {
		id string
		ts time.Time
	}
	entries := make([]entry, 0, len(batch))
	for id, ts := range batch {
		entries = append(entries, entry{id: id, ts: ts})
	}

	for i := 0; i < len(entries); i += batchChunkSize {
		chunk := entries[i:min(i+batchChunkSize, len(entries))]

		var (
			ids  []string
			args []any
			sb   strings.Builder
		)

		sb.WriteString("UPDATE devices SET last_seen = CASE id ")
		for _, e := range chunk {
			sb.WriteString("WHEN ? THEN GREATEST(last_seen, ?) ")
			args = append(args, e.id, e.ts)
			ids = append(ids, e.id)
		}
		sb.WriteString("END WHERE id IN (?)")
		args = append(args, ids)

		if err := r.db.WithContext(ctx).Exec(sb.String(), args...).Error; err != nil {
			return fmt.Errorf("failed to set last_seen batch (chunk %d): %w", i/batchChunkSize, err)
		}
	}

	return nil
}

func (r *Repository) Remove(ctx context.Context, filter ...SelectFilter) error {
	if len(filter) == 0 {
		return ErrInvalidFilter
	}

	f := newFilter(filter...)
	return f.apply(r.db.WithContext(ctx)).Delete(new(DeviceModel)).Error
}

func (r *Repository) Cleanup(ctx context.Context, until time.Time) (int64, error) {
	res := r.db.
		WithContext(ctx).
		Where("last_seen < ?", until).
		Delete(new(DeviceModel))

	return res.RowsAffected, res.Error
}
