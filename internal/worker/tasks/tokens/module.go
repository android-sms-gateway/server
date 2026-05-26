package tokens

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/jwt"
	"github.com/android-sms-gateway/server/internal/worker/executor"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"tokens",
		logger.WithNamedLogger("tokens"),
		fx.Provide(func(c Config) CleanupConfig {
			return c.Cleanup
		}, fx.Private),
		fx.Provide(jwt.NewRepository, fx.Private),
		fx.Provide(
			executor.AsWorkerTask(NewCleanupTask),
		),
	)
}
