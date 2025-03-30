package messages

import (
	"math"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/capcom6/go-helpers/slices"
)

func messageToDomain(input models.Message) MessageOut {
	var ttl *uint64 = nil
	if input.ValidUntil != nil {
		secondsUntil := uint64(math.Max(0, time.Until(*input.ValidUntil).Seconds()))
		ttl = &secondsUntil
	}

	return MessageOut{
		MessageIn: MessageIn{
			ID:                 input.ExtID,
			Message:            input.Message,
			PhoneNumbers:       slices.Map(input.Recipients, recipientToDomain),
			IsEncrypted:        input.IsEncrypted,
			SimNumber:          input.SimNumber,
			WithDeliveryReport: &input.WithDeliveryReport,
			TTL:                ttl,
			ValidUntil:         input.ValidUntil,
			Priority:           smsgateway.MessagePriority(input.Priority),
		},
		CreatedAt: input.CreatedAt,
	}
}

func recipientToDomain(input models.MessageRecipient) string {
	return input.PhoneNumber
}
