package settings

import (
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/settings"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type mobileControllerParams struct {
	fx.In

	DevicesSvc  *devices.Service
	SettingsSvc *settings.Service

	Logger *zap.Logger
}

type MobileController struct {
	base.Handler

	devicesSvc  *devices.Service
	settingsSvc *settings.Service
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
// Get settings
func (h *MobileController) get(device models.Device, c *fiber.Ctx) error {
	settings, err := h.settingsSvc.GetSettings(device.UserID, false)
	if err != nil {
		return fmt.Errorf("can't get settings for device %s (user ID: %s): %w", device.ID, device.UserID, err)
	}

	return c.JSON(settings)
}

func (h *MobileController) Register(router fiber.Router) {
	router.Get("", deviceauth.WithDevice(h.get))
}

func NewMobileController(params mobileControllerParams) *MobileController {
	return &MobileController{
		Handler: base.Handler{
			Logger: params.Logger.Named("settings"),
		},
		devicesSvc:  params.DevicesSvc,
		settingsSvc: params.SettingsSvc,
	}
}
