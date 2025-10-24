package push

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/cache"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/types"
	cacheImpl "github.com/android-sms-gateway/server/pkg/cache"
	"github.com/samber/lo"

	"go.uber.org/zap"
)

const (
	cachePrefixEvents    = "events:"
	cachePrefixBlacklist = "blacklist:"
)

type Config struct {
	Mode Mode

	ClientOptions map[string]string

	Debounce time.Duration
	Timeout  time.Duration
}

type Service struct {
	config Config

	client    client
	events    cache.Cache
	blacklist cache.Cache

	metrics *metrics
	logger  *zap.Logger
}

func New(
	config Config,
	client client,
	cacheFactory cache.Factory,
	metrics *metrics,
	logger *zap.Logger,
) (*Service, error) {
	events, err := cacheFactory.New(cachePrefixEvents)
	if err != nil {
		return nil, fmt.Errorf("can't create events cache: %w", err)
	}

	blacklist, err := cacheFactory.New(cachePrefixBlacklist)
	if err != nil {
		return nil, fmt.Errorf("can't create blacklist cache: %w", err)
	}

	config.Timeout = max(config.Timeout, time.Second)
	config.Debounce = max(config.Debounce, 5*time.Second)

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
func (s *Service) Enqueue(token string, event types.Event) error {
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
		return fmt.Errorf("can't serialize event wrapper: %w", err)
	}

	if err := s.events.Set(ctx, wrapper.key(), wrapperData); err != nil {
		s.metrics.IncError(1)
		return fmt.Errorf("can't add message to cache: %w", err)
	}

	s.metrics.IncEnqueued(string(event.Type))

	return nil
}

// sendAll sends messages to all targets from the cache after initializing the service.
func (s *Service) sendAll(ctx context.Context) {
	rawEvents, err := s.events.Drain(ctx)
	if err != nil {
		s.logger.Error("Can't drain cache", zap.Error(err))
		return
	}

	if len(rawEvents) == 0 {
		return
	}

	wrappers := lo.MapEntries(
		rawEvents,
		func(key string, value []byte) (string, *eventWrapper) {
			wrapper := new(eventWrapper)
			if err := wrapper.deserialize(value); err != nil {
				s.metrics.IncError(1)
				s.logger.Error("Failed to deserialize event wrapper", zap.String("key", key), zap.Binary("value", value), zap.Error(err))
				return "", nil
			}

			return wrapper.Token, wrapper
		},
	)
	delete(wrappers, "")

	messages := lo.MapValues(
		wrappers,
		func(value *eventWrapper, key string) Event {
			return value.Event
		},
	)

	s.logger.Info("Sending messages", zap.Int("count", len(messages)))
	sendCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	errs, err := s.client.Send(sendCtx, messages)
	if len(errs) == 0 && err == nil {
		s.logger.Info("Messages sent successfully", zap.Int("count", len(messages)))
		return
	}

	if err != nil {
		s.metrics.IncError(len(messages))
		s.logger.Error("Can't send messages", zap.Error(err))
		return
	}

	s.metrics.IncError(len(errs))

	for token, sendErr := range errs {
		s.logger.Error("Can't send message", zap.Error(sendErr), zap.String("token", token))

		wrapper := wrappers[token]
		wrapper.Retries++

		if wrapper.Retries >= maxRetries {
			if err := s.blacklist.Set(ctx, token, []byte{}, cacheImpl.WithTTL(blacklistTimeout)); err != nil {
				s.logger.Warn("Can't add to blacklist", zap.String("token", token), zap.Error(err))
			}

			s.metrics.IncBlacklist(BlacklistOperationAdded)
			s.metrics.IncRetry(RetryOutcomeMaxAttempts)
			s.logger.Warn("Retries exceeded, blacklisting token",
				zap.String("token", token),
				zap.Duration("ttl", blacklistTimeout))
			continue
		}

		wrapperData, err := wrapper.serialize()
		if err != nil {
			s.metrics.IncError(1)
			s.logger.Error("Can't serialize event wrapper", zap.Error(err))
			continue
		}

		if setErr := s.events.SetOrFail(ctx, wrapper.key(), wrapperData); setErr != nil {
			s.logger.Warn("Can't set message to cache", zap.Error(setErr))
			continue
		}

		s.metrics.IncRetry(RetryOutcomeRetried)
	}
}
