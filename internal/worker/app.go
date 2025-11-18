package worker

import (
	"context"

	"github.com/android-sms-gateway/server/internal/worker/config"
	"github.com/android-sms-gateway/server/internal/worker/executor"
	"github.com/android-sms-gateway/server/internal/worker/locker"
	"github.com/android-sms-gateway/server/internal/worker/server"
	"github.com/android-sms-gateway/server/internal/worker/tasks"
	"github.com/android-sms-gateway/server/pkg/health"
	"github.com/capcom6/go-infra-fx/db"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Run() {
	fx.New(
		logger.Module(),
		logger.WithFxDefaultLogger(),
		config.Module(),
		db.Module,
		fiberfx.Module(),
		module(),
	).Run()
}

func module() fx.Option {
	return fx.Module(
		"worker",
		locker.Module(),
		tasks.Module(),
		executor.Module(),
		health.Module(),
		server.Module(),
		fx.Invoke(func(logger *zap.Logger, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					logger.Info("worker started")
					return nil
				},
				OnStop: func(_ context.Context) error {
					logger.Info("worker stopped")
					return nil
				},
			})
		}),
	)
}
