package push

import (
	"errors"
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/client"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/fcm"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/upstream"
)

var ErrInvalidPushMode = errors.New("invalid push mode")

func newClient(config Config) (client.Client, error) {
	var (
		c   client.Client
		err error
	)

	switch config.Mode {
	case ModeFCM:
		c, err = fcm.New(config.ClientOptions)
	case ModeUpstream:
		c, err = upstream.New(config.ClientOptions)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidPushMode, config.Mode)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return c, nil
}
