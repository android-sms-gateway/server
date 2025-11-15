package push

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/client"
	cacheImpl "github.com/android-sms-gateway/server/pkg/cache"
	"github.com/samber/lo"

	"go.uber.org/zap"
)

const (
	cachePrefixEvents    = "events:"
	cachePrefixBlacklist = "blacklist:"

	defaultDebounce = 5 * time.Second
)

type Config struct {
	Mode Mode

	ClientOptions map[string]string

	Debounce time.Duration
	Timeout  time.Duration
}

type Service struct {
	config Config

	client    client.Client
	events    cache.Cache
	blacklist cache.Cache

	metrics *metrics
	logger  *zap.Logger
}

func New(
	config Config,
	client client.Client,
	cacheFactory cache.Factory,
	metrics *metrics,
	logger *zap.Logger,
) (*Service, error) {
	events, err := cacheFactory.New(cachePrefixEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to create events cache: %w", err)
	}

	blacklist, err := cacheFactory.New(cachePrefixBlacklist)
	if err != nil {
		return nil, fmt.Errorf("failed to create blacklist cache: %w", err)
	}

	config.Timeout = max(config.Timeout, time.Second)
	config.Debounce = max(config.Debounce, defaultDebounce)

	return &Service{
		config: config,

		client:    client,
		events:    events,
		blacklist: blacklist,

		metrics: metrics,
		logger:  logger,
	}, nil
}

// Run starts a ticker that triggers the sendAll function every debounce interval.
// It runs indefinitely until the provided context is canceled.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(s.config.Debounce)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendAll(ctx)
		}
	}
}

// Enqueue adds the data to the cache and immediately sends all messages if the debounce is 0.
func (s *Service) Enqueue(token string, event Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()

	if _, err := s.blacklist.Get(ctx, token); err == nil {
		s.metrics.IncBlacklist(BlacklistOperationSkipped)
		s.logger.Debug("Skipping blacklisted token", zap.String("token", token))
		return nil
	}

	wrapper := eventWrapper{
		Token:   token,
		Event:   event,
		Retries: 0,
	}
	wrapperData, err := wrapper.serialize()
	if err != nil {
		s.metrics.IncError(1)
		return fmt.Errorf("failed to serialize event wrapper: %w", err)
	}

	if setErr := s.events.Set(ctx, wrapper.key(), wrapperData); setErr != nil {
		s.metrics.IncError(1)
		return fmt.Errorf("failed to add message to cache: %w", setErr)
	}

	s.metrics.IncEnqueued(string(event.Type))

	return nil
}

// sendAll sends messages to all targets from the cache after initializing the service.
func (s *Service) sendAll(ctx context.Context) {
	rawEvents, err := s.events.Drain(ctx)
	if err != nil {
		s.logger.Error("failed to drain cache", zap.Error(err))
		return
	}

	if len(rawEvents) == 0 {
		return
	}

	wrappers := lo.FilterMap(
		lo.Values(rawEvents),
		func(value []byte, _ int) (*eventWrapper, bool) {
			wrapper := new(eventWrapper)
			if wrapErr := wrapper.deserialize(value); wrapErr != nil {
				s.metrics.IncError(1)
				s.logger.Error("failed to deserialize event wrapper", zap.Binary("value", value), zap.Error(wrapErr))
				return nil, false
			}

			return wrapper, true
		},
	)

	messages := lo.Map(
		wrappers,
		func(wrapper *eventWrapper, _ int) client.Message {
			return client.Message{
				Token: wrapper.Token,
				Event: wrapper.Event,
			}
		},
	)

	totalMessages := len(messages)
	if totalMessages == 0 {
		return
	}

	s.logger.Info("sending messages", zap.Int("total", totalMessages))

	sendCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()
	errs, err := s.client.Send(sendCtx, messages)
	if len(errs) == 0 && err == nil {
		s.logger.Info("messages sent successfully", zap.Int("total", totalMessages))
		return
	}

	if err != nil {
		s.metrics.IncError(totalMessages)
		s.logger.Error("failed to send messages", zap.Int("total", totalMessages), zap.Error(err))
		s.retry(ctx, wrappers)
		return
	}

	failed := lo.Filter(
		wrappers,
		func(item *eventWrapper, index int) bool {
			if sendErr := errs[index]; sendErr != nil {
				s.logger.Error("failed to send message", zap.String("token", item.Token), zap.Error(sendErr))
				return true
			}

			return false
		},
	)

	if len(failed) == 0 {
		return
	}

	s.metrics.IncError(len(failed))
	s.logger.Error("failed to send messages", zap.Int("total", totalMessages), zap.Int("failed", len(failed)))

	s.retry(ctx, failed)
}

func (s *Service) retry(ctx context.Context, events []*eventWrapper) {
	for _, wrapper := range events {
		token := wrapper.Token

		wrapper.Retries++

		if wrapper.Retries >= maxRetries {
			if err := s.blacklist.Set(ctx, token, []byte{}, cacheImpl.WithTTL(blacklistTimeout)); err != nil {
				s.logger.Warn("failed to blacklist", zap.String("token", token), zap.Error(err))
				continue
			}

			s.metrics.IncBlacklist(BlacklistOperationAdded)
			s.logger.Warn("retries exceeded, blacklisting token",
				zap.String("token", token),
				zap.Duration("ttl", blacklistTimeout),
			)
			continue
		}

		wrapperData, err := wrapper.serialize()
		if err != nil {
			s.metrics.IncError(1)
			s.logger.Error("failed to serialize event wrapper", zap.Error(err))
			continue
		}

		if setErr := s.events.SetOrFail(ctx, wrapper.key(), wrapperData); setErr != nil {
			s.logger.Warn("failed to set message to cache", zap.String("key", wrapper.key()), zap.Error(setErr))
			continue
		}

		s.metrics.IncRetry()
	}
}
