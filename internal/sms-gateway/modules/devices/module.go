package devices

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/cleaner"
	"github.com/capcom6/go-infra-fx/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type FxResult struct {
	fx.Out

	Service   *Service
	AsCleaner cleaner.Cleanable `group:"cleaners"`
}

var Module = fx.Module(
	"devices",
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("devices")
	}),
	fx.Provide(
		newDevicesRepository,
		fx.Private,
	),
	fx.Provide(func(p ServiceParams) FxResult {
		svc := NewService(p)
		return FxResult{
			Service:   svc,
			AsCleaner: svc,
		}
	}),
)

// init registers the Migrate function with the database migration system during package initialization.
func init() {
	db.RegisterMigration(Migrate)
}
