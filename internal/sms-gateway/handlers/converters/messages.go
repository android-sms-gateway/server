package converters

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
)

func MessageToDTO(m messages.MessageOut) smsgateway.MobileMessage {
	var message string
	var textMessage *smsgateway.TextMessage
	var dataMessage *smsgateway.DataMessage

	if m.TextContent != nil {
		message = m.TextContent.Text
		textMessage = &smsgateway.TextMessage{
			Text: m.TextContent.Text,
		}
	} else if m.DataContent != nil {
		dataMessage = &smsgateway.DataMessage{
			Data: m.DataContent.Data,
			Port: m.DataContent.Port,
		}
	}

	return smsgateway.MobileMessage{
		Message: smsgateway.Message{
			ID: m.ID,

			Message:     message,
			TextMessage: textMessage,
			DataMessage: dataMessage,

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
