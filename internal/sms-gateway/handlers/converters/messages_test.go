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

func TestMessageStateToDTO(t *testing.T) {
	// Set up a fixed time for testing
	now := time.Date(2023, 6, 15, 12, 30, 45, 123456789, time.UTC)

	// Define test cases
	tests := []struct {
		name     string
		input    messages.MessageState
		expected smsgateway.MessageState
	}{
		{
			name: "Full state with all fields",
			input: messages.MessageState{
				MessageStateInput: messages.MessageStateInput{
					ID:    "msg-123",
					State: messages.ProcessingStateDelivered,
					Recipients: []smsgateway.RecipientState{
						{PhoneNumber: "+1234567890", State: smsgateway.ProcessingStateDelivered},
						{
							PhoneNumber: "+9876543210",
							State:       smsgateway.ProcessingStateFailed,
							Error:       lo.ToPtr("timeout"),
						},
					},
					States: map[string]time.Time{
						string(messages.ProcessingStatePending):   now.Add(-time.Hour),
						string(messages.ProcessingStateDelivered): now,
					},
				},
				MessageStateContent: messages.MessageStateContent{
					MessageContent: messages.MessageContent{
						TextContent: &messages.TextMessageContent{Text: "Test message content"},
					},
				},
				DeviceID:    "device-1",
				IsHashed:    false,
				IsEncrypted: true,
				CreatedAt:   now,
			},
			expected: smsgateway.MessageState{
				ID:          "msg-123",
				DeviceID:    "device-1",
				State:       smsgateway.ProcessingStateDelivered,
				IsHashed:    false,
				IsEncrypted: true,
				Recipients: []smsgateway.RecipientState{
					{PhoneNumber: "+1234567890", State: smsgateway.ProcessingStateDelivered},
					{PhoneNumber: "+9876543210", State: smsgateway.ProcessingStateFailed, Error: lo.ToPtr("timeout")},
				},
				States: map[string]time.Time{
					string(messages.ProcessingStatePending):   now.Add(-time.Hour),
					string(messages.ProcessingStateDelivered): now,
				},
				TextMessage: &smsgateway.TextMessage{Text: "Test message content"},
				CreatedAt:   now,
			},
		},
		{
			name: "Data and hashed content",
			input: messages.MessageState{
				MessageStateInput: messages.MessageStateInput{
					ID:    "msg-456",
					State: messages.ProcessingStateSent,
					Recipients: []smsgateway.RecipientState{
						{PhoneNumber: "+1122334455", State: smsgateway.ProcessingStateSent},
					},
					States: map[string]time.Time{string(messages.ProcessingStateSent): now},
				},
				MessageStateContent: messages.MessageStateContent{
					MessageContent: messages.MessageContent{
						DataContent: &messages.DataMessageContent{Data: "SGVsbG8gV29ybGQh", Port: uint16(53739)},
					},
					HashedContent: &messages.HashedMessageContent{Hash: "1d4b6e3b1b6e3b1b6e3b1b6e3b1b6e3b1b6e3b1b"},
				},
				DeviceID:  "device-2",
				IsHashed:  true,
				CreatedAt: now.Add(time.Hour),
			},
			expected: smsgateway.MessageState{
				ID:          "msg-456",
				DeviceID:    "device-2",
				State:       smsgateway.ProcessingStateSent,
				IsHashed:    true,
				IsEncrypted: false,
				Recipients: []smsgateway.RecipientState{
					{PhoneNumber: "+1122334455", State: smsgateway.ProcessingStateSent},
				},
				States:      map[string]time.Time{string(messages.ProcessingStateSent): now},
				DataMessage: &smsgateway.DataMessage{Data: "SGVsbG8gV29ybGQh", Port: uint16(53739)},
				HashedMessage: &smsgateway.HashedMessage{
					Hash: "1d4b6e3b1b6e3b1b6e3b1b6e3b1b6e3b1b6e3b1b",
				},
				CreatedAt: now.Add(time.Hour),
			},
		},
		{
			name:     "Minimal state with zero values",
			input:    messages.MessageState{},
			expected: smsgateway.MessageState{},
		},
	}

	// Execute tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function under test
			result := converters.MessageStateToDTO(tc.input)

			// Assert the results
			assert.Equal(t, tc.expected, result)
		})
	}
}
