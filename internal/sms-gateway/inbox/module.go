package inbox

import (
	"github.com/capcom6/go-infra-fx/db"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

// Module returns the Fx module for the inbox messages feature.
func Module() fx.Option {
	return fx.Module(
		"inbox",
		logger.WithNamedLogger("inbox"),
		fx.Provide(NewRepository, fx.Private),
		fx.Provide(NewService),
	)
}

//nolint:gochecknoinits // framework-specific
func init() {
	db.RegisterMigration(Migrate)
}
