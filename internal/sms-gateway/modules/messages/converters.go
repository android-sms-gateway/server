package messages

import (
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/capcom6/go-helpers/anys"
	"github.com/capcom6/go-helpers/slices"
)

func messageToDTO(input *models.Message) smsgateway.Message {
	var ttl *uint64 = nil
	if input.ValidUntil != nil {
		delta := time.Until(*input.ValidUntil).Seconds()
		if delta > 0 {
			deltaInt := uint64(delta)
			ttl = &deltaInt
		} else {
			deltaInt := uint64(0)
			ttl = &deltaInt
		}
	}

	return smsgateway.Message{
		ID:           input.ExtID,
		Message:      input.Message,
		PhoneNumbers: slices.Map(input.Recipients, func(r models.MessageRecipient) string { return r.PhoneNumber }),

		SimNumber:          input.SimNumber,
		WithDeliveryReport: anys.AsPointer(input.WithDeliveryReport),
		IsEncrypted:        input.IsEncrypted,
		TTL:                ttl,
		ValidUntil:         input.ValidUntil,

		CreatedAt: input.CreatedAt,
	}
}
