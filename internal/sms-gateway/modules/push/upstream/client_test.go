package upstream_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/client"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/upstream"
)

func TestClientSendCarriesTokenType(t *testing.T) {
	t.Parallel()

	bodies := make(chan []byte, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bodies <- body
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c, err := upstream.New(map[string]string{"upstream_base_url": srv.URL})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err = c.Open(ctx); err != nil {
		t.Fatalf("failed to open client: %v", err)
	}
	defer func() { _ = c.Close(ctx) }()

	_, err = c.Send(ctx, []client.Message{
		{
			Token:     "device-fid",
			TokenType: client.TokenTypeFID,
			Event: client.Event{
				Type: smsgateway.PushMessageEnqueued,
				Data: map[string]string{"event": "MessageEnqueued"},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	var payload smsgateway.UpstreamPushRequest
	if err = json.Unmarshal(<-bodies, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if len(payload) != 1 {
		t.Fatalf("payload length = %d, want 1", len(payload))
	}
	if payload[0].PushTokenType == nil || *payload[0].PushTokenType != smsgateway.PushTokenTypeFID {
		t.Errorf("pushTokenType = %v, want %q", payload[0].PushTokenType, smsgateway.PushTokenTypeFID)
	}
}

func TestClientSendOmitsLegacyTokenType(t *testing.T) {
	t.Parallel()

	bodies := make(chan []byte, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bodies <- body
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c, err := upstream.New(map[string]string{"upstream_base_url": srv.URL})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err = c.Open(ctx); err != nil {
		t.Fatalf("failed to open client: %v", err)
	}
	defer func() { _ = c.Close(ctx) }()

	_, err = c.Send(ctx, []client.Message{
		{
			Token:     "legacy-token",
			TokenType: "",
			Event: client.Event{
				Type: smsgateway.PushMessageEnqueued,
				Data: nil,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	body := <-bodies
	if strings.Contains(string(body), `"pushTokenType"`) {
		t.Errorf("pushTokenType must be omitted for legacy tokens, got %s", body)
	}
}
