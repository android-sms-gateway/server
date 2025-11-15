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

type User struct {
	SoftDeletableModel

	ID           string   `gorm:"primaryKey;type:varchar(32)"`
	PasswordHash string   `gorm:"not null;type:varchar(72)"`
	Devices      []Device `gorm:"-,foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func NewUser(id, passwordHash string) *User {
	//nolint:exhaustruct // pertial constructor
	return &User{
		ID:           id,
		PasswordHash: passwordHash,
	}
}

type Device struct {
	SoftDeletableModel

	ID        string  `gorm:"primaryKey;type:char(21)"`
	Name      *string `gorm:"type:varchar(128)"`
	AuthToken string  `gorm:"not null;uniqueIndex;type:char(21)"`
	PushToken *string `gorm:"type:varchar(256)"`

	LastSeen time.Time `gorm:"not null;autocreatetime:false;default:CURRENT_TIMESTAMP(3);index:idx_devices_last_seen"`

	UserID string `gorm:"not null;type:varchar(32)"`
}

func NewDevice(name, pushToken *string) *Device {
	//nolint:exhaustruct // partial constructor
	return &Device{
		Name:      name,
		PushToken: pushToken,
	}
}

func (d *Device) IsEmpty() bool {
	if d == nil {
		return true
	}

	return d.ID == ""
}
