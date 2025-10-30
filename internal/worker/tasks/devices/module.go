package devices

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/worker/executor"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"devices",
		logger.WithNamedLogger("devices"),
		fx.Provide(func(c Config) CleanupConfig {
			return c.Cleanup
		}, fx.Private),
		fx.Provide(devices.NewRepository, fx.Private),
		fx.Provide(
			executor.AsWorkerTask(NewCleanupTask),
		),
	)
}
