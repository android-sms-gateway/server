package fcm

import (
	"context"
	"fmt"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/types"
	"google.golang.org/api/option"
)

type Client struct {
	options map[string]string

	client *messaging.Client
	mux    sync.Mutex
}

func New(options map[string]string) (*Client, error) {
	return &Client{
		options: options,
	}, nil
}

func (c *Client) Open(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.client != nil {
		return nil
	}

	creds := c.options["credentials"]
	if creds == "" {
		return fmt.Errorf("no credentials provided")
	}

	opt := option.WithCredentialsJSON([]byte(creds))

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("can't create firebase app: %w", err)
	}

	c.client, err = app.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("can't create firebase messaging client: %w", err)
	}

	return nil
}

func (c *Client) Send(ctx context.Context, messages map[string]types.Event) (map[string]error, error) {
	errs := make(map[string]error, len(messages))
	for address, payload := range messages {
		eventMap, err := eventToMap(payload)
		if err != nil {
			errs[address] = fmt.Errorf("can't marshal event: %w", err)
			continue
		}

		_, err = c.client.Send(ctx, &messaging.Message{
			Data: eventMap,
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
			Token: address,
		})

		if err != nil {
			errs[address] = fmt.Errorf("can't send message to %s: %w", address, err)
		}
	}

	return errs, nil
}

func (c *Client) Close(ctx context.Context) error {
	c.client = nil

	return nil
}
