package handlers

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/logs"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/messages"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/webhooks"
	"github.com/capcom6/go-infra-fx/http"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"handlers",
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("handlers")
	}),
	fx.Provide(
		http.AsRootHandler(newRootHandler),
		http.AsApiHandler(newThirdPartyHandler),
		http.AsApiHandler(newMobileHandler),
		http.AsApiHandler(newUpstreamHandler),
	),
	fx.Provide(
		newHealthHandler,
		messages.NewThirdPartyController,
		webhooks.NewThirdPartyController,
		webhooks.NewMobileController,
		devices.NewThirdPartyController,
		logs.NewThirdPartyController,
		fx.Private,
	),
)
