package devices

import "time"

type DeviceInput struct {
	DeviceInfo

	ID     string
	UserID string

	AuthToken string `json:"-"`
}

type DeviceInfo struct {
	DeviceUpdate

	Name *string
}

type DeviceUpdate struct {
	PushToken *string
	SimCards  []SimCard
}

type Device struct {
	DeviceInput

	LastSeen  time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (d Device) IsEmpty() bool {
	return d.ID == ""
}

type SimCard struct {
	SlotIndex   int // Zero-based index of the physical SIM slot (0, 1, ...).
	SimNumber   int // One-based number used by the application.
	PhoneNumber *string
	CarrierName *string
	ICCID       *string
}
