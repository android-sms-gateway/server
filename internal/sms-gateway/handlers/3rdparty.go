package handlers

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/logs"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/messages"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/jwtauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/settings"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/thirdparty"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/webhooks"
	"github.com/android-sms-gateway/server/internal/sms-gateway/jwt"
	"github.com/android-sms-gateway/server/internal/sms-gateway/users"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type thirdPartyHandler struct {
	base.Handler

	usersSvc *users.Service
	jwtSvc   jwt.Service

	healthHandler   *HealthHandler
	messagesHandler *messages.ThirdPartyController
	webhooksHandler *webhooks.ThirdPartyController
	devicesHandler  *devices.ThirdPartyController
	settingsHandler *settings.ThirdPartyController
	logsHandler     *logs.ThirdPartyController
	authHandler     *thirdparty.AuthHandler
}

func newThirdPartyHandler(
	usersSvc *users.Service,
	jwtService jwt.Service,

	healthHandler *HealthHandler,
	messagesHandler *messages.ThirdPartyController,
	webhooksHandler *webhooks.ThirdPartyController,
	devicesHandler *devices.ThirdPartyController,
	settingsHandler *settings.ThirdPartyController,
	logsHandler *logs.ThirdPartyController,
	authHandler *thirdparty.AuthHandler,

	logger *zap.Logger,
	validator *validator.Validate,
) *thirdPartyHandler {
	return &thirdPartyHandler{
		Handler: base.Handler{
			Logger:    logger,
			Validator: validator,
		},

		usersSvc: usersSvc,
		jwtSvc:   jwtService,

		healthHandler:   healthHandler,
		messagesHandler: messagesHandler,
		webhooksHandler: webhooksHandler,
		devicesHandler:  devicesHandler,
		settingsHandler: settingsHandler,
		logsHandler:     logsHandler,
		authHandler:     authHandler,
	}
}

func (h *thirdPartyHandler) Register(router fiber.Router) {
	router = router.Group("/3rdparty/v1")

	h.healthHandler.Register(router)

	router.Use(
		userauth.NewBasic(h.usersSvc),
		jwtauth.NewJWT(h.jwtSvc),
		userauth.UserRequired(),
	)

	h.authHandler.Register(router.Group("/auth"))

	h.messagesHandler.Register(router.Group("/message")) // TODO: remove after 2025-12-31
	h.messagesHandler.Register(router.Group("/messages"))

	h.devicesHandler.Register(router.Group("/device")) // TODO: remove after 2025-07-11
	h.devicesHandler.Register(router.Group("/devices"))

	h.settingsHandler.Register(router.Group("/settings"))

	h.webhooksHandler.Register(router.Group("/webhooks"))

	h.logsHandler.Register(router.Group("/logs"))
}
