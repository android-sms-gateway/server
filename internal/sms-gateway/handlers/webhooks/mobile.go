package webhooks

import (
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/webhooks"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type MobileController struct {
	base.Handler

	webhooksSvc *webhooks.Service
}

func NewMobileController(
	webhooksSvc *webhooks.Service,
	logger *zap.Logger,
	validator *validator.Validate,
) *MobileController {
	return &MobileController{
		Handler: base.Handler{
			Logger:    logger,
			Validator: validator,
		},
		webhooksSvc: webhooksSvc,
	}
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
// List webhooks.
func (h *MobileController) get(device models.Device, c *fiber.Ctx) error {
	items, err := h.webhooksSvc.Select(device.UserID, webhooks.WithDeviceID(device.ID, false))
	if err != nil {
		return fmt.Errorf("failed to select webhooks: %w", err)
	}

	return c.JSON(items)
}

func (h *MobileController) Register(router fiber.Router) {
	router.Get("", deviceauth.WithDevice(h.get))
}
