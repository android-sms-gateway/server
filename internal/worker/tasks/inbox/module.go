package inbox

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/inbox"
	"github.com/android-sms-gateway/server/internal/worker/executor"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"inbox",
		logger.WithNamedLogger("inbox"),
		fx.Provide(func(c Config) CleanupConfig {
			return c.Cleanup
		}, fx.Private),
		fx.Provide(inbox.NewRepository, fx.Private),
		fx.Provide(
			executor.AsWorkerTask(NewCleanupTask),
		),
	)
}
