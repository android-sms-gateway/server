package messages

import (
	"fmt"
	"math"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/capcom6/go-helpers/slices"
)

func messageToDomain(input messageModel) (Message, error) {
	var ttl *uint64
	if input.ValidUntil != nil {
		secondsUntil := uint64(math.Max(0, time.Until(*input.ValidUntil).Seconds()))
		ttl = &secondsUntil
	}

	textContent, err := input.GetTextContent()
	if err != nil {
		return Message{}, fmt.Errorf("failed to get text content: %w", err)
	}
	dataContent, err := input.GetDataContent()
	if err != nil {
		return Message{}, fmt.Errorf("failed to get data content: %w", err)
	}

	return Message{
		MessageInput: MessageInput{
			MessageContent: MessageContent{
				TextContent: textContent,
				DataContent: dataContent,
			},

			ID: input.ExtID,

			PhoneNumbers:       slices.Map(input.Recipients, recipientToDomain),
			IsEncrypted:        input.IsEncrypted,
			SimNumber:          input.SimNumber,
			WithDeliveryReport: &input.WithDeliveryReport,
			TTL:                ttl,
			ValidUntil:         input.ValidUntil,
			ScheduleAt:         input.ScheduleAt,
			Priority:           smsgateway.MessagePriority(input.Priority),
		},
		CreatedAt: input.CreatedAt,
	}, nil
}

func recipientToDomain(input messageRecipientModel) string {
	return input.PhoneNumber
}
