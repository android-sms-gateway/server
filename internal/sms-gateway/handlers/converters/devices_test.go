package converters_test

import (
	"testing"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/converters"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/capcom6/go-helpers/anys"
	"github.com/go-playground/assert/v2"
)

func TestDeviceToDTO(t *testing.T) {
	createdAt := time.Now()
	updatedAt := time.Now()
	lastSeenAt := time.Now()

	tests := []struct {
		name     string
		device   models.Device
		expected smsgateway.Device
	}{
		{
			name:     "empty device",
			device:   models.Device{},
			expected: smsgateway.Device{},
		},
		{
			name: "non-empty device",
			device: models.Device{
				ID:       "test-id",
				Name:     anys.AsPointer("test-name"),
				LastSeen: lastSeenAt,
				SoftDeletableModel: models.SoftDeletableModel{
					TimedModel: models.TimedModel{
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					},
				},
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
			device: models.Device{
				ID:   "test-id",
				Name: nil,
			},
			expected: smsgateway.Device{
				ID:   "test-id",
				Name: "",
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
