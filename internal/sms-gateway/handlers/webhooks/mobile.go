package webhooks

import (
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/webhooks"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type mobileControllerParams struct {
	fx.In

	WebhooksServices *webhooks.Service

	Logger *zap.Logger
}

type MobileController struct {
	base.Handler

	webhooksSvc *webhooks.Service
}

//	@Summary		List webhooks
//	@Description	Returns list of registered webhooks for device
//	@Security		MobileToken
//	@Tags			Device, Webhooks
//	@Produce		json
//	@Success		200	{object}	[]smsgateway.Webhook		"Webhook list"
//	@Failure		401	{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		500	{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/mobile/v1/webhooks [get]
//
// List webhooks
func (h *MobileController) get(device models.Device, c *fiber.Ctx) error {
	items, err := h.webhooksSvc.Select(device.UserID)
	if err != nil {
		return fmt.Errorf("can't select webhooks: %w", err)
	}

	return c.JSON(items)
}

func (h *MobileController) Register(router fiber.Router) {
	router.Get("", deviceauth.WithDevice(h.get))
}

func NewMobileController(params mobileControllerParams) *MobileController {
	return &MobileController{
		Handler: base.Handler{
			Logger: params.Logger.Named("webhooks"),
		},
		webhooksSvc: params.WebhooksServices,
	}
}
