package pubsub

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisConfig configures the Redis pubsub backend.
type RedisConfig struct {
	// Client is the Redis client to use.
	// If nil, a client is created from the URL.
	// If both Client and URL are provided, Client takes precedence.
	Client *redis.Client

	// URL is the Redis URL to use.
	// If empty, the Redis client is not created.
	URL string

	// Prefix is the prefix to use for all topics.
	Prefix string
}

type redisPubSub struct {
	prefix     string
	bufferSize uint

	client      *redis.Client
	ownedClient bool

	wg          sync.WaitGroup
	mu          sync.Mutex
	subscribers map[string]context.CancelFunc
	closeCh     chan struct{}
}

func NewRedis(config RedisConfig, opts ...Option) (*redisPubSub, error) {
	if config.Prefix != "" && !strings.HasSuffix(config.Prefix, ":") {
		config.Prefix += ":"
	}

	if config.Client == nil && config.URL == "" {
		return nil, fmt.Errorf("no redis client or url provided")
	}

	client := config.Client
	if client == nil {
		opt, err := redis.ParseURL(config.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis url: %w", err)
		}

		client = redis.NewClient(opt)
	}

	o := options{
		bufferSize: 0,
	}
	o.apply(opts...)

	return &redisPubSub{
		prefix:     config.Prefix,
		bufferSize: o.bufferSize,

		client:      client,
		ownedClient: config.Client == nil,

		subscribers: make(map[string]context.CancelFunc),
		closeCh:     make(chan struct{}),
	}, nil
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
				if msg == nil {
					continue
				}

				select {
				case ch <- Message{
					Topic: topic,
					Data:  []byte(msg.Payload),
				}:
				case <-r.closeCh:
					return
				case <-subCtx.Done():
					return
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

	if r.ownedClient {
		return r.client.Close()
	}

	return nil
}

var _ PubSub = (*redisPubSub)(nil)
