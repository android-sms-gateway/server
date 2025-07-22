package events

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/sse"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type MobileController struct {
	base.Handler

	sseSvc *sse.Service
}

func NewMobileController(sseService *sse.Service, validator *validator.Validate, logger *zap.Logger) *MobileController {
	return &MobileController{
		Handler: base.Handler{
			Logger:    logger,
			Validator: validator,
		},
		sseSvc: sseService,
	}
}

func (h *MobileController) get(device models.Device, c *fiber.Ctx) error {
	return h.sseSvc.Handler(device.ID, c)
}

func (h *MobileController) Register(router fiber.Router) {
	router.Get("", deviceauth.WithDevice(h.get))
}
