package messages

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type hashingWorker struct {
	interval time.Duration

	messages *Repository
	logger   *zap.Logger

	queue map[uint64]struct{}
	mux   sync.Mutex
}

func newHashingWorker(config Config, messages *Repository, logger *zap.Logger) *hashingWorker {
	return &hashingWorker{
		interval: config.HashingInterval,

		messages: messages,
		logger:   logger,

		queue: map[uint64]struct{}{},
	}
}

func (t *hashingWorker) Run(ctx context.Context) {
	t.logger.Info("Starting hashing task...")
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.logger.Info("Stopping hashing task...")
			return
		case <-ticker.C:
			t.process(ctx)
		}
	}
}

// Enqueue adds a message ID to the processing queue to be hashed in the next batch
func (t *hashingWorker) Enqueue(id uint64) {
	t.mux.Lock()
	t.queue[id] = struct{}{}
	t.mux.Unlock()
}

func (t *hashingWorker) process(ctx context.Context) {
	t.mux.Lock()

	ids := maps.Keys(t.queue)
	maps.Clear(t.queue)

	t.mux.Unlock()

	if len(ids) == 0 {
		return
	}

	t.logger.Debug("Hashing messages...")
	if _, err := t.messages.HashProcessed(ctx, ids); err != nil {
		t.logger.Error("Can't hash messages", zap.Error(err))
	}
}
