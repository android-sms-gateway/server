package push

import (
	"encoding/json"
	"fmt"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/client"
)

type Mode string

const (
	ModeFCM      Mode = "fcm"
	ModeUpstream Mode = "upstream"
)

type Event = client.Event

func normalizeTokenType(tokenType smsgateway.PushTokenType) smsgateway.PushTokenType {
	if tokenType == "" {
		return smsgateway.PushTokenTypeToken
	}

	return tokenType
}

type eventWrapper struct {
	Token     string                   `json:"token"`
	TokenType smsgateway.PushTokenType `json:"tokenType,omitempty"`
	Event     Event                    `json:"event"`
	Retries   int                      `json:"retries"`
}

func (e *eventWrapper) key() string {
	return string(normalizeTokenType(e.TokenType)) + ":" + e.Token + ":" + string(e.Event.Type)
}

func (e *eventWrapper) serialize() ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	return data, nil
}

func (e *eventWrapper) deserialize(data []byte) error {
	if err := json.Unmarshal(data, e); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	return nil
}
