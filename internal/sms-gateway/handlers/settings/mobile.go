package settings

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/settings"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type MobileController struct {
	base.Handler

	devicesSvc  *devices.Service
	settingsSvc *settings.Service
}

func NewMobileController(
	devicesSvc *devices.Service,
	settingsSvc *settings.Service,
	logger *zap.Logger,
	validator *validator.Validate,
) *MobileController {
	return &MobileController{
		Handler: base.Handler{
			Logger:    logger,
			Validator: validator,
		},
		devicesSvc:  devicesSvc,
		settingsSvc: settingsSvc,
	}
}

//	@Summary		Get settings
//	@Description	Returns settings for a device
//	@Security		MobileToken
//	@Tags			Device, Settings
//	@Produce		json
//	@Success		200	{object}	smsgateway.DeviceSettings	"Settings"
//	@Failure		401	{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		500	{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/mobile/v1/settings [get]
//
// Get settings.
func (h *MobileController) get(device models.Device, c *fiber.Ctx) error {
	settings, err := h.settingsSvc.GetSettings(device.UserID, false)
	if err != nil {
		h.Logger.Error(
			"failed to get settings",
			zap.Error(err),
			zap.String("device_id", device.ID),
			zap.String("user_id", device.UserID),
		)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get settings")
	}

	return c.JSON(settings)
}

func (h *MobileController) Register(router fiber.Router) {
	router.Get("", deviceauth.WithDevice(h.get))
}
