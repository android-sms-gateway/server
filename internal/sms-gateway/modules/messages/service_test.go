package messages

import (
	"reflect"
	"testing"

	"github.com/android-sms-gateway/client-go/smsgateway"
)

func TestService_recipientsStateToModel(t *testing.T) {
	type args struct {
		input []smsgateway.RecipientState
		hash  bool
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want []MessageRecipient
	}{
		{
			name: "Without +",
			s:    &Service{},
			args: args{
				input: []smsgateway.RecipientState{
					{
						PhoneNumber: "79990001234",
						State:       "",
					},
				},
			},
			want: []MessageRecipient{
				{
					MessageID:   0,
					PhoneNumber: "+79990001234",
					State:       "",
				},
			},
		},
		{
			name: "With +",
			s:    &Service{},
			args: args{
				input: []smsgateway.RecipientState{
					{
						PhoneNumber: "+79990001234",
						State:       "",
					},
				},
			},
			want: []MessageRecipient{
				{
					MessageID:   0,
					PhoneNumber: "+79990001234",
					State:       "",
				},
			},
		},
		{
			name: "With hashing",
			s:    &Service{},
			args: args{
				input: []smsgateway.RecipientState{
					{
						PhoneNumber: "+79990001234",
						State:       "",
					},
				},
				hash: true,
			},
			want: []MessageRecipient{
				{
					MessageID:   0,
					PhoneNumber: "62d17792b45c5307",
					State:       "",
				},
			},
		},
		{
			name: "Empty phone",
			s:    &Service{},
			args: args{
				input: []smsgateway.RecipientState{
					{
						PhoneNumber: "",
						State:       "",
					},
				},
			},
			want: []MessageRecipient{
				{
					MessageID:   0,
					PhoneNumber: "",
					State:       "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.recipientsStateToModel(tt.args.input, tt.args.hash); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessagesService.recipientsStateToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanPhoneNumber(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "Valid number with validation",
			input:       "+79161234567",
			expected:    "+79161234567",
			expectError: false,
		},
		{
			name:        "Invalid number with validation",
			input:       "+123!@#",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Empty input with validation",
			input:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Long number with validation",
			input:       "+345906566798696",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cleanPhoneNumber(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}
