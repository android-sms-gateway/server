package settings

import (
	"github.com/capcom6/go-infra-fx/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"settings",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("settings")
		}),
		fx.Provide(
			newRepository,
			fx.Private,
		),
		fx.Provide(
			NewService,
		),
	)
}

//nolint:gochecknoinits //backward compatibility
func init() {
	db.RegisterMigration(Migrate)
}
