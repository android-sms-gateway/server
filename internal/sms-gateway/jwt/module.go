package jwt

import (
	"github.com/capcom6/go-infra-fx/db"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"jwt",
		logger.WithNamedLogger("jwt"),
		fx.Provide(NewMetrics, NewRepository, fx.Private),
		fx.Provide(func(config Config, tokens *Repository, metrics *Metrics) (Service, error) {
			if config.Secret == "" {
				return newDisabled(), nil
			}

			return New(config, tokens, metrics)
		}),
	)
}

func init() {
	db.RegisterMigration(Migrate)
}
