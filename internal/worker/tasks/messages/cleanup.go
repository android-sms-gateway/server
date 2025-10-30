package messages

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/android-sms-gateway/server/internal/worker/executor"
	"go.uber.org/zap"
)

type cleanupTask struct {
	config   CleanupConfig
	messages *messages.Repository

	logger *zap.Logger
}

func NewCleanupTask(
	config CleanupConfig,
	messages *messages.Repository,
	logger *zap.Logger,
) executor.PeriodicTask {
	return &cleanupTask{
		config:   config,
		messages: messages,

		logger: logger,
	}
}

// Interval implements executor.PeriodicTask.
func (c *cleanupTask) Interval() time.Duration {
	return c.config.Interval
}

// Name implements executor.PeriodicTask.
func (c *cleanupTask) Name() string {
	return "messages:cleanup"
}

// Run implements executor.PeriodicTask.
func (c *cleanupTask) Run(ctx context.Context) error {
	rows, err := c.messages.Cleanup(ctx, time.Now().Add(-c.config.MaxAge))
	if err != nil {
		return fmt.Errorf("failed to cleanup messages: %w", err)
	}

	if rows > 0 {
		c.logger.Info("cleaned up messages", zap.Int64("rows", rows))
	}

	return nil
}

var _ executor.PeriodicTask = (*cleanupTask)(nil)
