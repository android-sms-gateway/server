package events

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/sse"
	"github.com/gofiber/fiber/v2"
)

type MobileController struct {
	base.Handler

	sseSvc *sse.Service
}

func NewMobileController(sseService *sse.Service) *MobileController {
	return &MobileController{
		sseSvc: sseService,
	}
}

func (h *MobileController) get(device models.Device, c *fiber.Ctx) error {
	return h.sseSvc.Handler(device.ID, c)
}

func (h *MobileController) Register(router fiber.Router) {
	router.Get("", deviceauth.WithDevice(h.get))
}
