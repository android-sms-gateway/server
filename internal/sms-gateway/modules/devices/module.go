package devices

import (
	"github.com/capcom6/go-infra-fx/db"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"devices",
		logger.WithNamedLogger("devices"),
		fx.Provide(
			NewRepository,
			fx.Private,
		),
		fx.Provide(NewService),
	)
}

//nolint:gochecknoinits // framework-specific
func init() {
	db.RegisterMigration(Migrate)
}
