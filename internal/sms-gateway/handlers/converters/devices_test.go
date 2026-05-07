package converters_test

import (
	"testing"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/converters"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/go-playground/assert/v2"
	"github.com/samber/lo"
)

func TestDeviceToDTO(t *testing.T) {
	createdAt := time.Now()
	updatedAt := time.Now()
	lastSeenAt := time.Now()

	tests := []struct {
		name     string
		device   devices.Device
		expected smsgateway.Device
	}{
		{
			name:     "empty device",
			device:   devices.Device{},
			expected: smsgateway.Device{},
		},
		{
			name: "non-empty device",
			device: devices.Device{
				DeviceInput: devices.DeviceInput{
					DeviceInfo: devices.DeviceInfo{
						DeviceUpdate: devices.DeviceUpdate{},
						Name:         lo.ToPtr("test-name"),
					},
					ID: "test-id",
				},
				LastSeen:  lastSeenAt,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			expected: smsgateway.Device{
				ID:        "test-id",
				Name:      "test-name",
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				LastSeen:  lastSeenAt,
			},
		},
		{
			name: "device with nil name",
			device: devices.Device{
				DeviceInput: devices.DeviceInput{
					DeviceInfo: devices.DeviceInfo{
						Name: nil,
					},
					ID: "test-id",
				},
			},
			expected: smsgateway.Device{
				ID:   "test-id",
				Name: "",
			},
		},
		{
			name: "device with sim cards",
			device: devices.Device{
				DeviceInput: devices.DeviceInput{
					DeviceInfo: devices.DeviceInfo{
						DeviceUpdate: devices.DeviceUpdate{
							SimCards: []devices.SimCard{
								{
									SlotIndex:   0,
									SimNumber:   1,
									PhoneNumber: lo.ToPtr("+79990001234"),
									CarrierName: lo.ToPtr("Carrier"),
									ICCID:       lo.ToPtr("8901260000000000000"),
								},
							},
						},
					},
					ID: "test-id",
				},
			},
			expected: smsgateway.Device{
				ID: "test-id",
				SimCards: []smsgateway.SimCard{
					{
						SlotIndex:   0,
						SimNumber:   1,
						PhoneNumber: lo.ToPtr("+79990001234"),
						CarrierName: lo.ToPtr("Carrier"),
						ICCID:       lo.ToPtr("8901260000000000000"),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := converters.DeviceToDTO(test.device)
			assert.Equal(t, test.expected, actual)
		})
	}
}
