package fcm

import (
	"context"
	"fmt"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/client"
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
		client:  nil,
		mux:     sync.Mutex{},
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
		return fmt.Errorf("%w: no credentials provided", ErrInitializationFailed)
	}

	opt := option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(creds))

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("%w: failed to create firebase app: %w", ErrInitializationFailed, err)
	}

	c.client, err = app.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to create firebase messaging client: %w", ErrInitializationFailed, err)
	}

	return nil
}

func (c *Client) Send(ctx context.Context, messages []client.Message) ([]error, error) {
	errs := make([]error, len(messages))

	for i, message := range messages {
		data, err := eventToMap(message.Event)
		if err != nil {
			errs[i] = fmt.Errorf("failed to marshal event: %w", err)
			continue
		}

		_, err = c.client.Send(ctx, newMessage(data, message.Token, message.TokenType))
		if err != nil {
			errs[i] = fmt.Errorf("failed to send message: %w", err)
		}
	}

	return errs, nil
}

func newMessage(data map[string]string, token string, tokenType smsgateway.PushTokenType) *messaging.Message {
	message := &messaging.Message{
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	}
	if tokenType == smsgateway.PushTokenTypeFID {
		message.Fid = token
	} else {
		message.Token = token //nolint:staticcheck // legacy registration tokens are still valid during the FCM transition
	}

	return message
}

func (c *Client) Close(_ context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.client = nil

	return nil
}
