package converters_test

import (
	"testing"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/converters"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/go-playground/assert/v2"
	"github.com/samber/lo"
)

func TestMessageToDTO(t *testing.T) {
	// Set up a fixed time for testing
	now := time.Now().UTC()

	// Define test cases
	tests := []struct {
		name     string
		input    messages.Message
		expected smsgateway.MobileMessage
	}{
		{
			name: "Full message with all fields",
			input: messages.Message{
				MessageInput: messages.MessageInput{
					MessageContent: messages.MessageContent{
						TextContent: &messages.TextMessageContent{Text: "Test message content"},
					},

					ID:                 "msg-123",
					PhoneNumbers:       []string{"+1234567890", "+9876543210"},
					IsEncrypted:        true,
					SimNumber:          lo.ToPtr(uint8(2)),
					WithDeliveryReport: lo.ToPtr(true),
					TTL:                lo.ToPtr(uint64(3600)),
					ValidUntil:         lo.ToPtr(now.Add(24 * time.Hour)),
					Priority:           100,
				},
				CreatedAt: now,
			},
			expected: smsgateway.MobileMessage{
				Message: smsgateway.Message{
					ID:                 "msg-123",
					Message:            "Test message content",
					TextMessage:        &smsgateway.TextMessage{Text: "Test message content"},
					PhoneNumbers:       []string{"+1234567890", "+9876543210"},
					IsEncrypted:        true,
					SimNumber:          lo.ToPtr(uint8(2)),
					WithDeliveryReport: lo.ToPtr(true),
					TTL:                lo.ToPtr(uint64(3600)),
					ValidUntil:         lo.ToPtr(now.Add(24 * time.Hour)),
					Priority:           100,
				},
				CreatedAt: now,
			},
		},
		{
			name: "Minimal message with required fields only",
			input: messages.Message{
				MessageInput: messages.MessageInput{
					MessageContent: messages.MessageContent{
						TextContent: &messages.TextMessageContent{Text: "Another test message"},
					},

					ID:           "msg-456",
					PhoneNumbers: []string{"+1122334455"},
				},
				CreatedAt: now,
			},
			expected: smsgateway.MobileMessage{
				Message: smsgateway.Message{
					ID:           "msg-456",
					Message:      "Another test message",
					TextMessage:  &smsgateway.TextMessage{Text: "Another test message"},
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
			result := converters.MessageToMobileDTO(tc.input)

			// Assert the results
			assert.Equal(t, tc.expected, result)
		})
	}
}
