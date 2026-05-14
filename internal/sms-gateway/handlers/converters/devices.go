package converters

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/capcom6/go-helpers/anys"
	"github.com/samber/lo"
)

func DeviceToDTO(device devices.Device) smsgateway.Device {
	return smsgateway.Device{
		ID:        device.ID,
		Name:      anys.OrDefault(device.Name, ""),
		CreatedAt: device.CreatedAt,
		UpdatedAt: device.UpdatedAt,
		DeletedAt: device.DeletedAt,
		LastSeen:  device.LastSeen,
		SimCards:  mapSimCards(device.SimCards),
	}
}

func mapSimCards(simCards []devices.SimCard) []smsgateway.SimCard {
	if simCards == nil {
		return nil
	}

	return lo.Map(simCards, func(sc devices.SimCard, _ int) smsgateway.SimCard {
		return smsgateway.SimCard{
			SlotIndex:   sc.SlotIndex,
			SimNumber:   sc.SimNumber,
			PhoneNumber: sc.PhoneNumber,
			CarrierName: sc.CarrierName,
			ICCID:       sc.ICCID,
		}
	})
}
