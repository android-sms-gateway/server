package tokens

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/jwt"
	"github.com/android-sms-gateway/server/internal/worker/executor"
	"go.uber.org/zap"
)

type cleanupTask struct {
	config CleanupConfig
	tokens *jwt.Repository

	logger *zap.Logger
}

func NewCleanupTask(
	config CleanupConfig,
	tokens *jwt.Repository,
	logger *zap.Logger,
) executor.PeriodicTask {
	return &cleanupTask{
		config: config,
		tokens: tokens,

		logger: logger,
	}
}

func (c *cleanupTask) Interval() time.Duration {
	return c.config.Interval
}

func (c *cleanupTask) Name() string {
	return "tokens:cleanup"
}

func (c *cleanupTask) Run(ctx context.Context) error {
	rows, err := c.tokens.Cleanup(ctx, time.Now().Add(-c.config.MaxAge))
	if err != nil {
		return fmt.Errorf("failed to cleanup tokens: %w", err)
	}

	if rows > 0 {
		c.logger.Info("cleaned up tokens", zap.Int64("rows", rows))
	}

	return nil
}

var _ executor.PeriodicTask = (*cleanupTask)(nil)
