package push

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/domain"
	"github.com/capcom6/go-helpers/cache"
	"github.com/capcom6/go-helpers/maps"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

	Client client

	Logger *zap.Logger
}

type Service struct {
	config Config

	client client

	cache     *cache.Cache[eventWrapper]
	blacklist *cache.Cache[struct{}]

	enqueuedCounter  *prometheus.CounterVec
	retriesCounter   *prometheus.CounterVec
	blacklistCounter *prometheus.CounterVec

	logger *zap.Logger
}

func New(params Params) *Service {
	if params.Config.Timeout == 0 {
		params.Config.Timeout = time.Second
	}
	if params.Config.Debounce < 5*time.Second {
		params.Config.Debounce = 5 * time.Second
	}

	enqueuedCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sms",
		Subsystem: "push",
		Name:      "enqueued_total",
		Help:      "Total number of messages enqueued",
	}, []string{"event"})

	retriesCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sms",
		Subsystem: "push",
		Name:      "retries_total",
		Help:      "Total retry attempts",
	}, []string{"outcome"})

	blacklistCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sms",
		Subsystem: "push",
		Name:      "blacklist_total",
		Help:      "Blacklist operations",
	}, []string{"operation"})

	return &Service{
		config: params.Config,
		client: params.Client,

		cache: cache.New[eventWrapper](cache.Config{}),
		blacklist: cache.New[struct{}](cache.Config{
			TTL: blacklistTimeout,
		}),

		enqueuedCounter:  enqueuedCounter,
		retriesCounter:   retriesCounter,
		blacklistCounter: blacklistCounter,

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
func (s *Service) Enqueue(token string, event *domain.Event) error {
	if _, err := s.blacklist.Get(token); err == nil {
		s.blacklistCounter.WithLabelValues(string(BlacklistOperationSkipped)).Inc()
		s.logger.Debug("Skipping blacklisted token", zap.String("token", token))
		return nil
	}

	wrapper := eventWrapper{
		token:   token,
		event:   event,
		retries: 0,
	}

	if err := s.cache.Set(token, wrapper); err != nil {
		return fmt.Errorf("can't add message to cache: %w", err)
	}

	s.enqueuedCounter.WithLabelValues(string(event.Event())).Inc()

	return nil
}

// sendAll sends messages to all targets from the cache after initializing the service.
func (s *Service) sendAll(ctx context.Context) {
	targets := s.cache.Drain()
	if len(targets) == 0 {
		return
	}

	messages := maps.MapValues(targets, func(w eventWrapper) domain.Event {
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
		s.logger.Error("Can't send messages", zap.Error(err))
		return
	}

	for token, sendErr := range errs {
		s.logger.Error("Can't send message", zap.Error(sendErr), zap.String("token", token))

		wrapper := targets[token]
		wrapper.retries++

		if wrapper.retries >= maxRetries {
			if err := s.blacklist.Set(token, struct{}{}); err != nil {
				s.logger.Warn("Can't add to blacklist", zap.String("token", token), zap.Error(err))
			}

			s.blacklistCounter.WithLabelValues(string(BlacklistOperationAdded)).Inc()
			s.retriesCounter.WithLabelValues(string(RetryOutcomeMaxAttempts)).Inc()
			s.logger.Warn("Retries exceeded, blacklisting token",
				zap.String("token", token),
				zap.Duration("ttl", blacklistTimeout))
			continue
		}

		if setErr := s.cache.SetOrFail(token, wrapper); setErr != nil {
			s.logger.Info("Can't set message to cache", zap.Error(setErr))
		}

		s.retriesCounter.WithLabelValues(string(RetryOutcomeRetried)).Inc()
	}
}
