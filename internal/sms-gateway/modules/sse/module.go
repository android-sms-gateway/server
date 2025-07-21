package sse

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"sse",
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("sse")
	}),
	fx.Provide(NewService),
)
