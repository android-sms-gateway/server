package devices

import (
	"errors"
	"fmt"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/converters"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type ThirdPartyController struct {
	base.Handler

	devicesSvc *devices.Service
}

func NewThirdPartyController(
	devicesSvc *devices.Service,
	logger *zap.Logger,
	validator *validator.Validate,
) *ThirdPartyController {
	return &ThirdPartyController{
		Handler: base.Handler{
			Logger:    logger,
			Validator: validator,
		},
		devicesSvc: devicesSvc,
	}
}

//	@Summary		List devices
//	@Description	Returns list of registered devices
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Devices
//	@Produce		json
//	@Success		200	{object}	[]smsgateway.Device			"Device list"
//	@Failure		400	{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		500	{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/3rdparty/v1/devices [get]
//
// List devices.
func (h *ThirdPartyController) get(userID string, c *fiber.Ctx) error {
	items, err := h.devicesSvc.Select(c.Context(), userID)
	if err != nil {
		return fmt.Errorf("failed to select devices: %w", err)
	}

	return c.JSON(lo.Map(
		items,
		func(device devices.Device, _ int) smsgateway.Device {
			return converters.DeviceToDTO(device)
		},
	))
}

//	@Summary		Remove device
//	@Description	Removes device
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Devices
//	@Produce		json
//	@Param			id	path	string	true	"Device ID"
//	@Success		204	"Successfully removed"
//	@Failure		400	{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		401	{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		404	{object}	smsgateway.ErrorResponse	"Device not found"
//	@Failure		500	{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/3rdparty/v1/devices/{id} [delete]
//
// Remove device.
func (h *ThirdPartyController) remove(userID string, c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.devicesSvc.Remove(c.Context(), userID, devices.WithID(id))
	if errors.Is(err, devices.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	if err != nil {
		return fmt.Errorf("failed to remove device: %w", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ThirdPartyController) Register(router fiber.Router) {
	router.Get("", permissions.RequireScope(ScopeList), userauth.WithUserID(h.get))
	router.Delete(":id", permissions.RequireScope(ScopeDelete), userauth.WithUserID(h.remove))
}
