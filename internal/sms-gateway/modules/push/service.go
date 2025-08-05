package push

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/types"
	"github.com/capcom6/go-helpers/cache"
	"github.com/capcom6/go-helpers/maps"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Config struct {
	Mode Mode

	ClientOptions map[string]string

	Debounce time.Duration
	Timeout  time.Duration
}

type Params struct {
	fx.In

	Config Config

	Client  client
	Metrics *metrics

	Logger *zap.Logger
}

type Service struct {
	config Config

	client  client
	metrics *metrics

	cache     *cache.Cache[eventWrapper]
	blacklist *cache.Cache[struct{}]

	logger *zap.Logger
}

func New(params Params) *Service {
	if params.Config.Timeout == 0 {
		params.Config.Timeout = time.Second
	}
	if params.Config.Debounce < 5*time.Second {
		params.Config.Debounce = 5 * time.Second
	}

	return &Service{
		config: params.Config,

		client:  params.Client,
		metrics: params.Metrics,

		cache: cache.New[eventWrapper](cache.Config{}),
		blacklist: cache.New[struct{}](cache.Config{
			TTL: blacklistTimeout,
		}),

		logger: params.Logger,
	}
}

// Run runs the service with the provided context if a debounce is set.
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
	if _, err := s.blacklist.Get(token); err == nil {
		s.metrics.IncBlacklist(BlacklistOperationSkipped)
		s.logger.Debug("Skipping blacklisted token", zap.String("token", token))
		return nil
	}

	wrapper := eventWrapper{
		token:   token,
		event:   &event,
		retries: 0,
	}

	if err := s.cache.Set(token, wrapper); err != nil {
		return fmt.Errorf("can't add message to cache: %w", err)
	}

	s.metrics.IncEnqueued(string(event.Type))

	return nil
}

// sendAll sends messages to all targets from the cache after initializing the service.
func (s *Service) sendAll(ctx context.Context) {
	targets := s.cache.Drain()
	if len(targets) == 0 {
		return
	}

	messages := maps.MapValues(targets, func(w eventWrapper) types.Event {
		return *w.event
	})

	s.logger.Info("Sending messages", zap.Int("count", len(messages)))
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	errs, err := s.client.Send(ctx, messages)
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

		wrapper := targets[token]
		wrapper.retries++

		if wrapper.retries >= maxRetries {
			if err := s.blacklist.Set(token, struct{}{}); err != nil {
				s.logger.Warn("Can't add to blacklist", zap.String("token", token), zap.Error(err))
			}

			s.metrics.IncBlacklist(BlacklistOperationAdded)
			s.metrics.IncRetry(RetryOutcomeMaxAttempts)
			s.logger.Warn("Retries exceeded, blacklisting token",
				zap.String("token", token),
				zap.Duration("ttl", blacklistTimeout))
			continue
		}

		if setErr := s.cache.SetOrFail(token, wrapper); setErr != nil {
			s.logger.Info("Can't set message to cache", zap.Error(setErr))
		}

		s.metrics.IncRetry(RetryOutcomeRetried)
	}
}
