package locker

import (
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
	)
}
