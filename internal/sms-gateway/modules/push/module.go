package push

import (
	"context"
	"errors"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/fcm"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/upstream"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"push",
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("push")
	}),
	fx.Provide(newMetrics, fx.Private),
	fx.Provide(
		func(cfg Config, lc fx.Lifecycle) (c client, err error) {
			switch cfg.Mode {
			case ModeFCM:
				c, err = fcm.New(cfg.ClientOptions)
			case ModeUpstream:
				c, err = upstream.New(cfg.ClientOptions)
			default:
				return nil, errors.New("invalid push mode")
			}

			if err != nil {
				return nil, err
			}

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return c.Open(ctx)
				},
				OnStop: func(ctx context.Context) error {
					return c.Close(ctx)
				},
			})

			return c, nil
		},
		fx.Private,
	),
	fx.Provide(
		New,
	),
)
