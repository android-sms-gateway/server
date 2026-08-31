package converters

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
)

func MessageToMobileDTO(m messages.Message) smsgateway.MobileMessage {
	var message string
	var textMessage *smsgateway.TextMessage
	var dataMessage *smsgateway.DataMessage
	var mmsMessage *smsgateway.MmsMessage

	switch {
	case m.TextContent != nil:
		message = m.TextContent.Text
		textMessage = m.TextContent
	case m.DataContent != nil:
		dataMessage = m.DataContent
	case m.MmsContent != nil:
		mmsMessage = m.MmsContent
	}

	return smsgateway.MobileMessage{
		Message: smsgateway.Message{
			ID:       m.ID,
			DeviceID: "",

			Message:     message,
			TextMessage: textMessage,
			DataMessage: dataMessage,
			MmsMessage:  mmsMessage,

			SimNumber:          m.SimNumber,
			WithDeliveryReport: m.WithDeliveryReport,
			IsEncrypted:        m.IsEncrypted,
			PhoneNumbers:       m.PhoneNumbers,
			TTL:                m.TTL,
			ValidUntil:         m.ValidUntil,
			ScheduleAt:         m.ScheduleAt,
			Priority:           m.Priority,
		},
		State:     smsgateway.ProcessingState(m.State),
		CreatedAt: m.CreatedAt,
	}
}

func MessageStateToDTO(state messages.MessageState) smsgateway.MessageState {
	return smsgateway.MessageState{
		ID:          state.ID,
		DeviceID:    state.DeviceID,
		State:       smsgateway.ProcessingState(state.State),
		IsHashed:    state.IsHashed,
		IsEncrypted: state.IsEncrypted,
		Recipients:  state.Recipients,
		States:      state.States,

		TextMessage:   state.TextContent,
		DataMessage:   state.DataContent,
		MmsMessage:    state.MmsContent,
		HashedMessage: state.HashedContent,
	}
}
