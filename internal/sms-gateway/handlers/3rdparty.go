package handlers

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/logs"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/messages"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/webhooks"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/auth"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ThirdPartyHandlerParams struct {
	fx.In

	HealthHandler   *healthHandler
	MessagesHandler *messages.ThirdPartyController
	WebhooksHandler *webhooks.ThirdPartyController
	DevicesHandler  *devices.ThirdPartyController
	LogsHandler     *logs.ThirdPartyController

	AuthSvc *auth.Service

	Logger    *zap.Logger
	Validator *validator.Validate
}

type thirdPartyHandler struct {
	base.Handler

	healthHandler   *healthHandler
	messagesHandler *messages.ThirdPartyController
	webhooksHandler *webhooks.ThirdPartyController
	devicesHandler  *devices.ThirdPartyController
	logsHandler     *logs.ThirdPartyController

	authSvc *auth.Service
}

func (h *thirdPartyHandler) Register(router fiber.Router) {
	router = router.Group("/3rdparty/v1")

	h.healthHandler.Register(router)

	router.Use(
		userauth.New(h.authSvc),
		userauth.UserRequired(),
	)

	h.messagesHandler.Register(router.Group("/message")) // TODO: remove after 2025-12-31
	h.messagesHandler.Register(router.Group("/messages"))

	h.devicesHandler.Register(router.Group("/device")) // TODO: remove after 2025-07-11
	h.devicesHandler.Register(router.Group("/devices"))

	h.webhooksHandler.Register(router.Group("/webhooks"))

	h.logsHandler.Register(router.Group("/logs"))
}

func newThirdPartyHandler(params ThirdPartyHandlerParams) *thirdPartyHandler {
	return &thirdPartyHandler{
		Handler:         base.Handler{Logger: params.Logger.Named("ThirdPartyHandler"), Validator: params.Validator},
		healthHandler:   params.HealthHandler,
		messagesHandler: params.MessagesHandler,
		webhooksHandler: params.WebhooksHandler,
		devicesHandler:  params.DevicesHandler,
		logsHandler:     params.LogsHandler,
		authSvc:         params.AuthSvc,
	}
}
