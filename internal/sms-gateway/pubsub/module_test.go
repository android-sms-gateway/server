package pubsub_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apppubsub "github.com/android-sms-gateway/server/internal/sms-gateway/pubsub"
	"github.com/go-core-fx/pubsubfx"
	"github.com/go-core-fx/pubsubfx/pubsub"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// receiveOne returns the next message or fails the test if none arrives
// before the timeout.
func receiveOne(t *testing.T, sub *pubsub.Subscription) pubsub.Message {
	t.Helper()

	select {
	case msg, ok := <-sub.Receive():
		if !ok {
			t.Fatal("subscription channel closed before a message arrived")
		}

		return msg
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for a message")

		return pubsub.Message{}
	}
}

// TestModuleProvidesPubSub boots the wrapper together with the pubsubfx module
// the way the app wires them, and checks a publish/subscribe round trip
// resolves through the container and Stop closes the instance.
func TestModuleProvidesPubSub(t *testing.T) {
	t.Parallel()

	var ps pubsub.PubSub

	app := fx.New(
		fx.Supply(zap.NewNop()),
		fx.Supply(apppubsub.Config{URL: "memory://", BufferSize: 4}),
		apppubsub.Module(),
		pubsubfx.Module(),
		fx.Invoke(func(p pubsub.PubSub) {
			ps = p
		}),
	)
	if app.Err() != nil {
		t.Fatalf("fx.New failed: %v", app.Err())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	sub, err := ps.Subscribe(ctx, "events")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	if pubErr := ps.Publish(ctx, "events", []byte("hello")); pubErr != nil {
		t.Fatalf("Publish failed: %v", pubErr)
	}

	if msg := receiveOne(t, sub); string(msg.Data) != "hello" {
		t.Errorf("Expected data %q, got %q", "hello", msg.Data)
	}

	if stopErr := app.Stop(ctx); stopErr != nil {
		t.Fatalf("stop failed: %v", stopErr)
	}

	// pubsubfx.Module()'s OnStop hook must have closed the instance.
	if pubErr := ps.Publish(context.Background(), "events", []byte("x")); !errors.Is(pubErr, pubsub.ErrPubSubClosed) {
		t.Errorf("Expected ErrPubSubClosed after Stop, got %v", pubErr)
	}
}
