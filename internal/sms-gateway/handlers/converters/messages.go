package converters

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
)

func MessageToDTO(m messages.MessageOut) smsgateway.MobileMessage {
	return smsgateway.MobileMessage{
		Message: smsgateway.Message{
			ID:                 m.ID,
			Message:            m.Message,
			SimNumber:          m.SimNumber,
			WithDeliveryReport: m.WithDeliveryReport,
			IsEncrypted:        m.IsEncrypted,
			PhoneNumbers:       m.PhoneNumbers,
			TTL:                m.TTL,
			ValidUntil:         m.ValidUntil,
			Priority:           m.Priority,
		},
		CreatedAt: m.CreatedAt,
	}
}
