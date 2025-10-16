package messages

import (
	"context"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type HashingTaskConfig struct {
	Interval time.Duration
}

type HashingTaskParams struct {
	fx.In

	Messages *repository
	Config   HashingTaskConfig
	Logger   *zap.Logger
}

type hashingTask struct {
	Messages *repository
	Config   HashingTaskConfig
	Logger   *zap.Logger

	queue map[uint64]struct{}
	mux   sync.Mutex
}

func (t *hashingTask) Run(ctx context.Context) {
	t.Logger.Info("Starting hashing task...")
	ticker := time.NewTicker(t.Config.Interval)
	defer ticker.Stop()

	t.Logger.Info("Initial hashing...")
	if err := t.Messages.HashProcessed([]uint64{}); err != nil {
		t.Logger.Error("Can't hash messages", zap.Error(err))
	}
	t.Logger.Info("Initial hashing...Done")

	for {
		select {
		case <-ctx.Done():
			t.Logger.Info("Stopping hashing task...")
			return
		case <-ticker.C:
			t.process()
		}
	}
}

// Enqueue adds a message ID to the processing queue to be hashed in the next batch
func (t *hashingTask) Enqueue(id uint64) {
	t.mux.Lock()
	t.queue[id] = struct{}{}
	t.mux.Unlock()
}

func (t *hashingTask) process() {
	t.mux.Lock()

	ids := maps.Keys(t.queue)
	maps.Clear(t.queue)

	t.mux.Unlock()

	if len(ids) == 0 {
		return
	}

	t.Logger.Debug("Hashing messages...")
	if err := t.Messages.HashProcessed(ids); err != nil {
		t.Logger.Error("Can't hash messages", zap.Error(err))
	}
}

func newHashingTask(params HashingTaskParams) *hashingTask {
	return &hashingTask{
		Messages: params.Messages,
		Config:   params.Config,
		Logger:   params.Logger,
		queue:    map[uint64]struct{}{},
	}
}
