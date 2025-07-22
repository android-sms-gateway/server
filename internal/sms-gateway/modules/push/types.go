package push

import (
	"context"

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
	Send(ctx context.Context, messages map[string]types.Event) (map[string]error, error)
	Close(ctx context.Context) error
}

type eventWrapper struct {
	token   string
	event   *types.Event
	retries int
}
