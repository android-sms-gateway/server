package push

import (
	"context"
	"encoding/json"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/types"
)

type Mode string

const (
	ModeFCM      Mode = "fcm"
	ModeUpstream Mode = "upstream"
)

type Event = types.Event

type client interface {
	Open(ctx context.Context) error
	Send(ctx context.Context, messages []types.Message) ([]error, error)
	Close(ctx context.Context) error
}

type eventWrapper struct {
	Token   string `json:"token"`
	Event   Event  `json:"event"`
	Retries int    `json:"retries"`
}

func (e *eventWrapper) key() string {
	return e.Token + ":" + string(e.Event.Type)
}

func (e *eventWrapper) serialize() ([]byte, error) {
	return json.Marshal(e)
}

func (e *eventWrapper) deserialize(data []byte) error {
	return json.Unmarshal(data, e)
}
