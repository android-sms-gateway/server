package messages

import (
	"math"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/capcom6/go-helpers/slices"
)

// messageToDomain converts a models.Message to a MessageOut domain object. It maps the input message’s fields—including the external ID, text, recipient phone numbers (converted via recipientToDomain), encryption flag, SIM number, delivery report indicator, and priority—to the corresponding fields in MessageOut. If the input specifies a ValidUntil timestamp, the function computes a non-negative TTL (time-to-live) in seconds relative to the current time and assigns it. The output also preserves the original creation timestamp.
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

// RecipientToDomain extracts and returns the phone number from the given MessageRecipient.
func recipientToDomain(input models.MessageRecipient) string {
	return input.PhoneNumber
}
