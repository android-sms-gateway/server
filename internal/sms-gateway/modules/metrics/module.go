//nolint:revive // ignore package name
package metrics

import (
	"github.com/capcom6/go-infra-fx/http"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"metrics",
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("metrics")
		}),
		fx.Provide(
			http.AsRootHandler(newHTTPHandler),
		),
	)
}
