package locker

import (
	"context"
	"database/sql"

	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	const timeoutSeconds = 10

	return fx.Module(
		"locker",
		logger.WithNamedLogger("locker"),
		fx.Provide(func(db *sql.DB) Locker {
			return NewMySQLLocker(db, "worker:", timeoutSeconds)
		}),
		fx.Invoke(func(locker Locker, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return nil
				},
				OnStop: func(_ context.Context) error {
					return locker.Close()
				},
			})
		}),
	)
}
