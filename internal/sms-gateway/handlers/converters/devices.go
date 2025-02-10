package converters

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/pkg/types"
)

func DeviceToDTO(device models.Device) smsgateway.Device {
	return smsgateway.Device{
		ID:        device.ID,
		Name:      types.OrDefault(device.Name, ""),
		CreatedAt: device.CreatedAt,
		UpdatedAt: device.UpdatedAt,
		DeletedAt: device.DeletedAt,
		LastSeen:  device.LastSeen,
	}
}
