package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/client"
	"github.com/samber/lo"
)

const baseURL = "https://api.sms-gate.app/upstream/v1"

var ErrInvalidResponse = errors.New("invalid response")

type Client struct {
	options map[string]string

	client *http.Client
	mux    sync.Mutex
}

func New(options map[string]string) (*Client, error) {
	return &Client{
		options: options,
		client:  nil,
		mux:     sync.Mutex{},
	}, nil
}

func (c *Client) Open(_ context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.client != nil {
		return nil
	}

	c.client = &http.Client{}

	return nil
}

func (c *Client) Send(ctx context.Context, messages []client.Message) ([]error, error) {
	payload := lo.Map(
		messages,
		func(item client.Message, _ int) smsgateway.PushNotification {
			return smsgateway.PushNotification{
				Token: item.Token,
				Event: item.Event.Type,
				Data:  item.Event.Data,
			}
		},
	)

	payloadBytes, err := json.Marshal(smsgateway.UpstreamPushRequest(payload)) //nolint:unconvert //type checking

	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/push", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "sms-gate/1.x (server; golang)")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		return c.mapErrors(
			messages,
			fmt.Errorf("%w: unexpected status code: %d", ErrInvalidResponse, resp.StatusCode),
		), nil
	}

	return nil, nil
}

func (c *Client) mapErrors(messages []client.Message, err error) []error {
	return lo.Map(
		messages,
		func(_ client.Message, _ int) error {
			return err
		},
	)
}

func (c *Client) Close(_ context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.client = nil

	return nil
}
