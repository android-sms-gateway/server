package health

import (
	"context"
	"errors"
	"fmt"
	"io"
	httpclient "net/http"
	"time"

	"github.com/capcom6/go-infra-fx/http"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var ErrNotHealthy = errors.New("not healthy")

type Checker struct {
	config http.Config

	shutdowner fx.Shutdowner
	logger     *zap.Logger
}

func NewChecker(config http.Config, shutdowner fx.Shutdowner, logger *zap.Logger) *Checker {
	return &Checker{
		config:     config,
		shutdowner: shutdowner,
		logger:     logger,
	}
}

func (c *Checker) Execute(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	client := httpclient.DefaultClient

	req, err := httpclient.NewRequestWithContext(
		ctx,
		httpclient.MethodGet,
		"http://"+c.config.Listen+"/health/live",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Info(string(body))

	if res.StatusCode >= httpclient.StatusBadRequest {
		c.logger.Error("health check failed", zap.Int("status", res.StatusCode), zap.String("body", string(body)))
		return fmt.Errorf("%w: health check failed: %s", ErrNotHealthy, string(body))
	}

	c.logger.Info("health check passed", zap.Int("status", res.StatusCode))

	if shErr := c.shutdowner.Shutdown(); shErr != nil {
		c.logger.Error("failed to shutdown", zap.Error(shErr))
	}

	return nil
}
