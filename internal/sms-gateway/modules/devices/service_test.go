package devices_test

import (
	"context"
	"testing"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/go-playground/assert/v2"
	"github.com/samber/lo"
)

func TestServiceInsertRejectsInconsistentE2E(t *testing.T) {
	tests := []struct {
		name       string
		publicKey  *string
		keyVersion *int
	}{
		{
			name:       "public key set without key version",
			publicKey:  lo.ToPtr("test-public-key"),
			keyVersion: nil,
		},
		{
			name:       "key version set without public key",
			publicKey:  nil,
			keyVersion: lo.ToPtr(1),
		},
		{
			name:       "empty public key with key version",
			publicKey:  lo.ToPtr(""),
			keyVersion: lo.ToPtr(1),
		},
		{
			name:       "empty public key without key version",
			publicKey:  lo.ToPtr(""),
			keyVersion: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &devices.Service{}

			_, err := service.Insert(context.Background(), "test-user", devices.DeviceInfo{
				DeviceUpdate: devices.DeviceUpdate{
					PublicKey:  test.publicKey,
					KeyVersion: test.keyVersion,
				},
			})

			assert.Equal(t, devices.ErrInconsistentE2E, err)
		})
	}
}

func TestValidateE2EPairRejects(t *testing.T) {
	tests := []struct {
		name       string
		publicKey  *string
		keyVersion *int
	}{
		{
			name:       "public key set without key version",
			publicKey:  lo.ToPtr("test-public-key"),
			keyVersion: nil,
		},
		{
			name:       "key version set without public key",
			publicKey:  nil,
			keyVersion: lo.ToPtr(1),
		},
		{
			name:       "empty public key with key version",
			publicKey:  lo.ToPtr(""),
			keyVersion: lo.ToPtr(1),
		},
		{
			name:       "empty public key without key version",
			publicKey:  lo.ToPtr(""),
			keyVersion: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, devices.ErrInconsistentE2E, devices.ValidateE2EPair(test.publicKey, test.keyVersion))
		})
	}
}

func TestValidateE2EPair(t *testing.T) {
	tests := []struct {
		name       string
		publicKey  *string
		keyVersion *int
	}{
		{
			name:       "both nil",
			publicKey:  nil,
			keyVersion: nil,
		},
		{
			name:       "both set",
			publicKey:  lo.ToPtr("test-public-key"),
			keyVersion: lo.ToPtr(1),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, nil, devices.ValidateE2EPair(test.publicKey, test.keyVersion))
		})
	}
}

func TestServiceValidateE2EPairRejects(t *testing.T) {
	tests := []struct {
		name       string
		publicKey  *string
		keyVersion *int
	}{
		{
			name:       "public key set without key version",
			publicKey:  lo.ToPtr("test-public-key"),
			keyVersion: nil,
		},
		{
			name:       "key version set without public key",
			publicKey:  nil,
			keyVersion: lo.ToPtr(1),
		},
		{
			name:       "empty public key with key version",
			publicKey:  lo.ToPtr(""),
			keyVersion: lo.ToPtr(1),
		},
		{
			name:       "empty public key without key version",
			publicKey:  lo.ToPtr(""),
			keyVersion: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &devices.Service{}

			err := service.ValidateE2EPair(test.publicKey, test.keyVersion)

			assert.Equal(t, devices.ErrInconsistentE2E, err)
		})
	}
}

func TestServiceValidateE2EPair(t *testing.T) {
	tests := []struct {
		name       string
		publicKey  *string
		keyVersion *int
	}{
		{
			name:       "both nil",
			publicKey:  nil,
			keyVersion: nil,
		},
		{
			name:       "both set",
			publicKey:  lo.ToPtr("test-public-key"),
			keyVersion: lo.ToPtr(1),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &devices.Service{}

			err := service.ValidateE2EPair(test.publicKey, test.keyVersion)

			assert.Equal(t, nil, err)
		})
	}
}
