package client

import (
	"context"

	"github.com/android-sms-gateway/client-go/smsgateway"
)

type Client interface {
	Open(ctx context.Context) error
	Send(ctx context.Context, messages []Message) ([]error, error)
	Close(ctx context.Context) error
}

const (
	TokenTypeFID   = "fid"
	TokenTypeToken = "token"
)

type Message struct {
	Token     string
	TokenType string
	Event     Event
}

type Event struct {
	Type smsgateway.PushEventType `json:"type"`
	Data map[string]string        `json:"data"`
}
