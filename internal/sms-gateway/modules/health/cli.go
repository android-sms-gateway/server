package health

import (
	"io"
	httpclient "net/http"
	"time"

	"github.com/capcom6/go-infra-fx/http"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func testHealth(shutdowner fx.Shutdowner, logger *zap.Logger, config http.Config) {
	client := httpclient.Client{
		Timeout: 1 * time.Second,
	}

	res, err := client.Get("http://" + config.Listen + "/health")
	if err != nil {
		logger.Error("Failed to send request", zap.Error(err))
		if err := shutdowner.Shutdown(fx.ExitCode(1)); err != nil {
			logger.Error("Failed to shutdown", zap.Error(err))
		}
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Error("Failed to close body", zap.Error(err))
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Error("Failed to read body", zap.Error(err))
	}

	logger.Info(string(body))

	if res.StatusCode >= 400 {
		if err := shutdowner.Shutdown(fx.ExitCode(1)); err != nil {
			logger.Error("Failed to shutdown", zap.Error(err))
		}
		return
	}

	if err := shutdowner.Shutdown(); err != nil {
		logger.Error("Failed to shutdown", zap.Error(err))
	}
}
