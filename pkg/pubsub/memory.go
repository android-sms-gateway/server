package pubsub

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type memoryPubSub struct {
	bufferSize uint

	wg      sync.WaitGroup
	mu      sync.RWMutex
	topics  map[string]map[string]chan Message
	closeCh chan struct{}
}

func NewMemory(opts ...Option) *memoryPubSub {
	o := options{
		bufferSize: 0,
	}
	o.apply(opts...)

	return &memoryPubSub{
		bufferSize: o.bufferSize,

		topics:  make(map[string]map[string]chan Message),
		closeCh: make(chan struct{}),
	}
}

// Publish sends a message to all subscribers of the given topic.
// This method blocks until all subscribers have received the message
// or until ctx is cancelled or the pubsub instance is closed.
func (m *memoryPubSub) Publish(ctx context.Context, topic string, data []byte) error {
	select {
	case <-m.closeCh:
		return ErrPubSubClosed
	default:
	}

	if topic == "" {
		return ErrInvalidTopic
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	subscribers, exists := m.topics[topic]
	if !exists {
		return nil
	}

	wg := &sync.WaitGroup{}
	msg := Message{Topic: topic, Data: data}

	for _, ch := range subscribers {
		wg.Add(1)
		go func(ch chan Message) {
			defer wg.Done()

			select {
			case ch <- msg:
			case <-ctx.Done():
				return
			case <-m.closeCh:
				return
			}
		}(ch)
	}

	wg.Wait()

	return nil
}

func (m *memoryPubSub) Subscribe(ctx context.Context, topic string) (*Subscription, error) {
	select {
	case <-m.closeCh:
		return nil, ErrPubSubClosed
	default:
	}

	if topic == "" {
		return nil, ErrInvalidTopic
	}

	id := uuid.NewString()
	subCtx, cancel := context.WithCancel(ctx)
	ch := make(chan Message, m.bufferSize)

	m.subscribe(id, topic, ch)

	m.wg.Add(1)
	go func() {
		select {
		case <-subCtx.Done():
		case <-m.closeCh:
		}

		cancel()
		m.unsubscribe(id, topic)
		close(ch)

		m.wg.Done()
	}()

	return &Subscription{id: id, ctx: subCtx, cancel: cancel, ch: ch}, nil
}

func (m *memoryPubSub) subscribe(id, topic string, ch chan Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscriptions, ok := m.topics[topic]
	if !ok {
		subscriptions = make(map[string]chan Message)
		m.topics[topic] = subscriptions
	}
	subscriptions[id] = ch
}

func (m *memoryPubSub) unsubscribe(id, topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscriptions, ok := m.topics[topic]
	if !ok {
		return
	}
	delete(subscriptions, id)
	if len(subscriptions) == 0 {
		delete(m.topics, topic)
	}
}

func (m *memoryPubSub) Close() error {
	select {
	case <-m.closeCh:
		return nil
	default:
	}
	close(m.closeCh)

	m.wg.Wait()

	return nil
}

var _ PubSub = (*memoryPubSub)(nil)
