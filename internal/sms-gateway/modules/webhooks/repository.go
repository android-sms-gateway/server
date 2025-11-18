package webhooks

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Select(filters ...SelectFilter) ([]*Webhook, error) {
	webhooks := []*Webhook{}
	if err := newFilter(filters...).apply(r.db).Find(&webhooks).Error; err != nil {
		return nil, err
	}
	return webhooks, nil
}

func (r *Repository) Replace(webhook *Webhook) error {
	return r.db.
		Clauses(clause.OnConflict{UpdateAll: true}).
		Save(webhook).
		Error
}

func (r *Repository) Delete(filters ...SelectFilter) error {
	return newFilter(filters...).apply(r.db).Delete(new(Webhook)).Error
}
