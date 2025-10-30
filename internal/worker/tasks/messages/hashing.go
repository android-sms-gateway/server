package messages

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/android-sms-gateway/server/internal/worker/executor"
	"go.uber.org/zap"
)

type initialHashingTask struct {
	config   HashingConfig
	messages *messages.Repository

	logger *zap.Logger
}

func NewInitialHashingTask(
	config HashingConfig,
	messages *messages.Repository,
	logger *zap.Logger,
) executor.PeriodicTask {
	return &initialHashingTask{
		config:   config,
		messages: messages,

		logger: logger,
	}
}

// Interval implements tasks.PeriodicTask.
func (i *initialHashingTask) Interval() time.Duration {
	return i.config.Interval
}

// Name implements tasks.PeriodicTask.
func (i *initialHashingTask) Name() string {
	return "messages:hashing"
}

// Run implements tasks.PeriodicTask.
func (i *initialHashingTask) Run(ctx context.Context) error {
	rows, err := i.messages.HashProcessed(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to hash processed messages: %w", err)
	}

	if rows > 0 {
		i.logger.Info("hashed messages", zap.Int64("rows", rows))
	}

	return nil
}

var _ executor.PeriodicTask = (*initialHashingTask)(nil)
