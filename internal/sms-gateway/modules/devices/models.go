package devices

import (
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/samber/lo"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DeviceModel struct {
	models.SoftDeletableModel

	ID        string  `gorm:"primaryKey;type:char(21)"`
	Name      *string `gorm:"type:varchar(128)"`
	AuthToken string  `gorm:"not null;uniqueIndex;type:char(21)"`
	PushToken *string `gorm:"type:varchar(256)"`

	LastSeen time.Time `gorm:"not null;autocreatetime:false;default:CURRENT_TIMESTAMP(3);index:idx_devices_last_seen"`

	UserID string `gorm:"not null;type:varchar(32)"`

	SimCards datatypes.JSONSlice[simCardModel] `gorm:"serializer:json;type:json"`
}

func newDeviceModel(device DeviceInput) *DeviceModel {
	now := time.Now()
	return &DeviceModel{
		SoftDeletableModel: models.SoftDeletableModel{
			TimedModel: models.TimedModel{
				CreatedAt: now,
				UpdatedAt: now,
			},
			DeletedAt: nil,
		},

		ID:        device.ID,
		Name:      device.Name,
		AuthToken: device.AuthToken,
		PushToken: device.PushToken,
		LastSeen:  now,
		UserID:    device.UserID,
		SimCards: lo.Map(
			device.SimCards,
			func(simCard SimCard, _ int) simCardModel { return newSimCardModel(simCard) },
		),
	}
}

func (*DeviceModel) TableName() string {
	return "devices"
}

func (m *DeviceModel) toDomain() *Device {
	if m == nil {
		return nil
	}

	return &Device{
		DeviceInput: DeviceInput{
			DeviceInfo: DeviceInfo{
				DeviceUpdate: DeviceUpdate{
					PushToken: m.PushToken,
					SimCards:  lo.Map(m.SimCards, func(m simCardModel, _ int) SimCard { return m.toDomain() }),
				},

				Name: m.Name,
			},

			ID:     m.ID,
			UserID: m.UserID,

			AuthToken: m.AuthToken,
		},

		LastSeen:  m.LastSeen,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}
}

type simCardModel struct {
	SlotIndex   int     `json:"slotIndex"`
	SimNumber   int     `json:"simNumber"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	CarrierName *string `json:"carrierName,omitempty"`
	ICCID       *string `json:"iccid,omitempty"`
}

func newSimCardModel(simCard SimCard) simCardModel {
	return simCardModel(simCard)
}

func (m simCardModel) toDomain() SimCard {
	return SimCard(m)
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(DeviceModel)); err != nil {
		return fmt.Errorf("devices migration failed: %w", err)
	}
	return nil
}
