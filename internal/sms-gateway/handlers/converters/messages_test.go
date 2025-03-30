package converters_test

import (
	"testing"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/converters"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/capcom6/go-helpers/anys"
	"github.com/go-playground/assert/v2"
)

func TestMessageToDTO(t *testing.T) {
	// Set up a fixed time for testing
	now := time.Now().UTC()

	// Define test cases
	tests := []struct {
		name     string
		input    messages.MessageOut
		expected smsgateway.MobileMessage
	}{
		{
			name: "Full message with all fields",
			input: messages.MessageOut{
				MessageIn: messages.MessageIn{
					ID:                 "msg-123",
					Message:            "Test message content",
					PhoneNumbers:       []string{"+1234567890", "+9876543210"},
					IsEncrypted:        true,
					SimNumber:          anys.AsPointer(uint8(2)),
					WithDeliveryReport: anys.AsPointer(true),
					TTL:                anys.AsPointer(uint64(3600)),
					ValidUntil:         anys.AsPointer(now.Add(24 * time.Hour)),
					Priority:           100,
				},
				CreatedAt: now,
			},
			expected: smsgateway.MobileMessage{
				Message: smsgateway.Message{
					ID:                 "msg-123",
					Message:            "Test message content",
					PhoneNumbers:       []string{"+1234567890", "+9876543210"},
					IsEncrypted:        true,
					SimNumber:          anys.AsPointer(uint8(2)),
					WithDeliveryReport: anys.AsPointer(true),
					TTL:                anys.AsPointer(uint64(3600)),
					ValidUntil:         anys.AsPointer(now.Add(24 * time.Hour)),
					Priority:           100,
				},
				CreatedAt: now,
			},
		},
		{
			name: "Minimal message with required fields only",
			input: messages.MessageOut{
				MessageIn: messages.MessageIn{
					ID:           "msg-456",
					Message:      "Another test message",
					PhoneNumbers: []string{"+1122334455"},
				},
				CreatedAt: now,
			},
			expected: smsgateway.MobileMessage{
				Message: smsgateway.Message{
					ID:           "msg-456",
					Message:      "Another test message",
					PhoneNumbers: []string{"+1122334455"},
				},
				CreatedAt: now,
			},
		},
	}

	// Execute tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function under test
			result := converters.MessageToDTO(tc.input)

			// Assert the results
			assert.Equal(t, tc.expected, result)
		})
	}
}
