package settings

import (
	"fmt"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type thirdPartyControllerParams struct {
	fx.In

	DevicesSvc *devices.Service

	Validator *validator.Validate
	Logger    *zap.Logger
}

type ThirdPartyController struct {
	base.Handler

	devicesSvc *devices.Service
}

// @Summary		Get settings
// @Description	Returns settings for a specific user
// @Security		ApiAuth
// @Tags			User, Settings
// @Produce		json
// @Success		200	{object}	smsgateway.DeviceSettings	"Settings"
// @Failure		401	{object}	smsgateway.ErrorResponse	"Unauthorized"
// @Failure		500	{object}	smsgateway.ErrorResponse	"Internal server error"
// @Router			/3rdparty/v1/settings [get]
//
// Get settings
func (h *ThirdPartyController) get(user models.User, c *fiber.Ctx) error {
	settings, err := h.devicesSvc.GetSettings(user.ID)
	if err != nil {
		return fmt.Errorf("can't get settings: %w", err)
	}

	return c.JSON(settings)
}

// @Summary		Update settings
// @Description	Updates settings for a specific user
// @Security		ApiAuth
// @Tags			User, Settings
// @Accept			json
// @Produce		json
// @Param			request	body		smsgateway.DeviceSettings	true	"Settings"
// @Success		200		{object}	object						"Settings updated"
// @Failure		400		{object}	smsgateway.ErrorResponse	"Invalid request"
// @Failure		401		{object}	smsgateway.ErrorResponse	"Unauthorized"
// @Failure		500		{object}	smsgateway.ErrorResponse	"Internal server error"
// @Router			/3rdparty/v1/settings [put]
//
// Update settings
func (h *ThirdPartyController) put(user models.User, c *fiber.Ctx) error {
	if err := h.BodyParserValidator(c, &smsgateway.DeviceSettings{}); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	settings := make(map[string]any, 8)

	if err := c.BodyParser(&settings); err != nil {
		return err
	}

	updated, err := h.devicesSvc.ReplaceSettings(user.ID, settings)

	if err != nil {
		return fmt.Errorf("can't update settings: %w", err)
	}

	return c.JSON(updated)
}

func (h *ThirdPartyController) patch(user models.User, c *fiber.Ctx) error {
	if err := h.BodyParserValidator(c, &smsgateway.DeviceSettings{}); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	settings := make(map[string]any, 8)

	if err := c.BodyParser(&settings); err != nil {
		return err
	}

	updated, err := h.devicesSvc.UpdateSettings(user.ID, settings)
	if err != nil {
		return fmt.Errorf("can't update settings: %w", err)
	}

	return c.JSON(updated)
}

func (h *ThirdPartyController) Register(app fiber.Router) {
	app.Get("", userauth.WithUser(h.get))
	app.Patch("", userauth.WithUser(h.patch))
	app.Put("", userauth.WithUser(h.put))
}

// NewThirdPartyController creates a new ThirdPartyController with the provided device service, validator, and logger.
func NewThirdPartyController(params thirdPartyControllerParams) *ThirdPartyController {
	return &ThirdPartyController{
		Handler: base.Handler{
			Logger:    params.Logger.Named("settings"),
			Validator: params.Validator,
		},
		devicesSvc: params.DevicesSvc,
	}
}
