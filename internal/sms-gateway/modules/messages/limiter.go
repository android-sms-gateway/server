package messages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/go-core-fx/cachefx/cache"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type limitItem struct {
	Reason string `json:"reason"`
}

func (i *limitItem) Marshal() ([]byte, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal limit item: %w", err)
	}

	return data, nil
}

func (i *limitItem) Unmarshal(data []byte) error {
	if err := json.Unmarshal(data, i); err != nil {
		return fmt.Errorf("failed to unmarshal limit item: %w", err)
	}

	return nil
}

type Limiter struct {
	config QueueConfig

	messages *Repository
	cache    *cache.Typed[*limitItem]

	refreshQueue map[string]struct{}
	mux          sync.Mutex

	metrics *metrics
	logger  *zap.Logger
}

func NewLimiter(
	config QueueConfig,
	messages *Repository,
	storage cache.Cache,
	metrics *metrics,
	logger *zap.Logger,
) *Limiter {
	return &Limiter{
		config: config,

		messages: messages,
		cache:    cache.NewTyped[*limitItem](storage),

		refreshQueue: make(map[string]struct{}),
		mux:          sync.Mutex{},

		metrics: metrics,
		logger:  logger,
	}
}

func (l *Limiter) Run(ctx context.Context) {
	l.logger.Info("Starting limiter...")
	ticker := time.NewTicker(l.config.StatsRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			l.logger.Info("Limiter stopped")
			return
		case <-ticker.C:
			l.process(ctx)
		}
	}
}

func (l *Limiter) Check(ctx context.Context, deviceID string) error {
	if l.config.IsEmpty() {
		return nil
	}

	item, err := l.cache.Get(ctx, l.makeKey(deviceID))
	if errors.Is(err, cache.ErrKeyNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to get limit item from cache: %w", err)
	}

	if item == nil || item.Reason == "" {
		l.metrics.IncLimiterCheck(false)
		return nil
	}

	l.logger.Warn("Queue limit exceeded", zap.String("device_id", deviceID), zap.String("reason", item.Reason))
	l.metrics.IncLimiterCheck(true)

	return fmt.Errorf("%w: %s", ErrQueueLimitExceeded, item.Reason)
}

func (l *Limiter) Refresh(_ context.Context, deviceID string) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.refreshQueue[deviceID] = struct{}{}

	l.logger.Debug("Queued device for refresh", zap.String("device_id", deviceID))
	l.metrics.IncLimiterRefresh()
	return nil
}

func (l *Limiter) process(ctx context.Context) {
	l.mux.Lock()
	ids := slices.AppendSeq(make([]string, 0, len(l.refreshQueue)), maps.Keys(l.refreshQueue))
	clear(l.refreshQueue)
	l.mux.Unlock()

	if len(ids) == 0 {
		return
	}

	l.logger.Debug("Refreshing queue limits", zap.Int("count", len(ids)))
	l.metrics.SetLimiterBatchSize(len(ids))

	for _, deviceID := range ids {
		item := l.processDevice(ctx, deviceID)
		if err := l.cache.Set(ctx, l.makeKey(deviceID), item, cache.WithTTL(l.config.StatsCacheTTL)); err != nil {
			l.logger.Error(
				"failed to cache limit item",
				zap.String("device_id", deviceID),
				zap.Any("item", item),
				zap.Error(err),
			)
		}
	}
}

func (l *Limiter) processDevice(ctx context.Context, deviceID string) *limitItem {
	item := l.checkOldestPending(ctx, deviceID)
	if item != nil {
		return item
	}

	item = l.checkMaxPending(ctx, deviceID)
	if item != nil {
		return item
	}

	item = l.checkLastStates(ctx, deviceID)
	if item != nil {
		return item
	}

	return item
}

func (l *Limiter) checkMaxPending(ctx context.Context, deviceID string) *limitItem {
	if l.config.MaxPending <= 0 {
		return nil
	}

	pendingCount, err := l.messages.CountPending(ctx, deviceID)
	if err != nil {
		l.logger.Error("failed to count pending messages", zap.String("device_id", deviceID), zap.Error(err))
		l.metrics.IncLimiterQueryError(checkMaxPending)
		return nil
	}

	if pendingCount >= l.config.MaxPending {
		return &limitItem{Reason: fmt.Sprintf("too many pending messages: %d / %d", pendingCount, l.config.MaxPending)}
	}

	return nil
}

func (l *Limiter) checkOldestPending(ctx context.Context, deviceID string) *limitItem {
	if l.config.MaxPendingAge <= 0 {
		return nil
	}

	oldestPending, err := l.messages.GetOldestPendingTime(ctx, deviceID)
	if err != nil {
		l.logger.Error("failed to get oldest pending message", zap.String("device_id", deviceID), zap.Error(err))
		l.metrics.IncLimiterQueryError(checkMaxPendingAge)
		return nil
	}

	if oldestPending != nil && time.Since(*oldestPending) > l.config.MaxPendingAge {
		return &limitItem{Reason: fmt.Sprintf("too old pending message: %s", oldestPending.Format(time.RFC3339))}
	}

	return nil
}

func (l *Limiter) checkLastStates(ctx context.Context, deviceID string) *limitItem {
	if l.config.MaxFailed <= 0 || l.config.MaxFailedAge <= 0 {
		return nil
	}

	since := time.Now().Add(-l.config.MaxFailedAge)
	lastStates, err := l.messages.GetStatesInTimeWindow(ctx, deviceID, since, l.config.MaxFailed)
	if err != nil {
		l.logger.Error("failed to get failed states in time window", zap.String("device_id", deviceID), zap.Error(err))
		l.metrics.IncLimiterQueryError(checkMaxFailed)
		return nil
	}

	if len(lastStates) < l.config.MaxFailed {
		return nil
	}

	if lo.EveryBy(
		lastStates,
		func(state ProcessingState) bool {
			return state == ProcessingStateFailed
		},
	) {
		return &limitItem{Reason: "too many failed messages"}
	}

	return nil
}

func (l *Limiter) makeKey(deviceID string) string {
	return "limit:" + deviceID
}
