package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisPubSub struct {
	prefix     string
	bufferSize uint

	client *redis.Client

	wg          sync.WaitGroup
	mu          sync.Mutex
	subscribers map[string]context.CancelFunc
	closeCh     chan struct{}
}

func NewRedis(client *redis.Client, prefix string, opts ...Option) *redisPubSub {
	o := options{
		bufferSize: 0,
	}
	o.apply(opts...)

	return &redisPubSub{
		prefix:     prefix,
		bufferSize: o.bufferSize,

		client: client,

		subscribers: make(map[string]context.CancelFunc),
		closeCh:     make(chan struct{}),
	}
}

func (r *redisPubSub) Publish(ctx context.Context, topic string, data []byte) error {
	select {
	case <-r.closeCh:
		return ErrPubSubClosed
	default:
	}

	if topic == "" {
		return ErrInvalidTopic
	}

	return r.client.Publish(ctx, r.prefix+topic, data).Err()
}

func (r *redisPubSub) Subscribe(ctx context.Context, topic string) (*Subscription, error) {
	select {
	case <-r.closeCh:
		return nil, ErrPubSubClosed
	default:
	}

	if topic == "" {
		return nil, ErrInvalidTopic
	}

	ps := r.client.Subscribe(ctx, r.prefix+topic)
	_, err := ps.Receive(ctx)
	if err != nil {
		closeErr := ps.Close()
		return nil, errors.Join(fmt.Errorf("can't subscribe: %w", err), closeErr)
	}

	id := uuid.NewString()
	subCtx, cancel := context.WithCancel(ctx)
	ch := make(chan Message, r.bufferSize)

	// Track this subscriber
	r.mu.Lock()
	r.subscribers[id] = cancel
	r.mu.Unlock()

	r.wg.Add(1)
	go func() {
		defer func() {
			_ = ps.Close()
			close(ch)

			r.mu.Lock()
			delete(r.subscribers, id)
			r.mu.Unlock()

			r.wg.Done()
		}()

		for {
			select {
			case <-r.closeCh:
				return
			case <-subCtx.Done():
				return
			case msg, ok := <-ps.Channel():
				if !ok {
					return
				}
				if msg != nil {
					ch <- Message{
						Topic: topic,
						Data:  []byte(msg.Payload),
					}
				}
			}
		}
	}()

	return &Subscription{id: id, ctx: subCtx, cancel: cancel, ch: ch}, nil
}

func (r *redisPubSub) Close() error {
	select {
	case <-r.closeCh:
		return nil
	default:
		close(r.closeCh)
	}

	r.wg.Wait()

	return nil
}

var _ PubSub = (*redisPubSub)(nil)
