package inbox

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/inbox"
	"github.com/android-sms-gateway/server/internal/worker/executor"
	"go.uber.org/zap"
)

type cleanupTask struct {
	config CleanupConfig
	inbox  *inbox.Repository

	logger *zap.Logger
}

func NewCleanupTask(
	config CleanupConfig,
	inbox *inbox.Repository,
	logger *zap.Logger,
) executor.PeriodicTask {
	return &cleanupTask{
		config: config,
		inbox:  inbox,
		logger: logger,
	}
}

// Interval implements executor.PeriodicTask.
func (c *cleanupTask) Interval() time.Duration {
	return c.config.Interval
}

// Name implements executor.PeriodicTask.
func (c *cleanupTask) Name() string {
	return "inbox:cleanup"
}

// Run implements executor.PeriodicTask.
func (c *cleanupTask) Run(ctx context.Context) error {
	rows, err := c.inbox.Cleanup(ctx, time.Now().Add(-c.config.MaxAge))
	if err != nil {
		return fmt.Errorf("failed to cleanup inbox messages: %w", err)
	}

	if rows > 0 {
		c.logger.Info("cleaned up inbox messages", zap.Int64("rows", rows))
	}

	return nil
}

var _ executor.PeriodicTask = (*cleanupTask)(nil)
