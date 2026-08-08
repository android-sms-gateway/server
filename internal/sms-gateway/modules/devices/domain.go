package devices

import (
	"time"
)

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
	// PublicKey is a base64-encoded RSA public key (nil if no E2E).
	// Setting a new key together with KeyVersion overwrites the previous key;
	// clearing an existing key is intentionally unsupported. On insert, nil
	// means the device is created without E2E; on update, a both-nil pair is
	// a no-op that leaves the existing key unchanged. An empty string is
	// rejected with ErrInconsistentE2E.
	PublicKey *string
	// KeyVersion is the key version used for rotation tracking (nil if no
	// E2E). It must always be set together with PublicKey: providing exactly
	// one of the two is rejected with ErrInconsistentE2E.
	KeyVersion *int
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
