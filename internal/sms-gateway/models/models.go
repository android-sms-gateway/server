package models

import (
	"time"
)

type TimedModel struct {
	CreatedAt time.Time `gorm:"->;not null;autocreatetime:false;default:CURRENT_TIMESTAMP(3)"`
	UpdatedAt time.Time `gorm:"->;not null;autoupdatetime:false;default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

type SoftDeletableModel struct {
	TimedModel

	DeletedAt *time.Time `gorm:"<-:update"`
}
