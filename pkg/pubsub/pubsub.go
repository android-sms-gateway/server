package pubsub

import (
	"context"
	"errors"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
	ErrPubSubClosed  = errors.New("pubsub is closed")
	ErrInvalidTopic  = errors.New("invalid topic name")
)

type Message struct {
	Topic string
	Data  []byte
}

type Subscription struct {
	id     string
	ch     <-chan Message
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Subscription) Receive() <-chan Message {
	return s.ch
}

func (s *Subscription) Close() {
	s.cancel()
}

type Subscriber interface {
	// Subscribe subscribes to a topic and returns a channel for receiving messages.
	// The channel will be closed when the context is cancelled.
	Subscribe(ctx context.Context, topic string) (*Subscription, error)
}

type Publisher interface {
	// Publish publishes a message to a topic.
	// All subscribers to the topic will receive the message (fan-out).
	Publish(ctx context.Context, topic string, data []byte) error
}

type PubSub interface {
	Publisher
	Subscriber
	// Close closes the pubsub instance and releases all resources.
	Close() error
}
