package handlers

import (
	"fmt"
	"strings"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/converters"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/events"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/messages"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/settings"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/webhooks"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/auth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/users"
	"github.com/capcom6/go-helpers/anys"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/jaevor/go-nanoid"
	"go.uber.org/zap"
)

type mobileHandler struct {
	base.Handler

	authSvc    *auth.Service
	usersSvc   *users.Service
	devicesSvc *devices.Service

	messagesCtrl *messages.MobileController
	webhooksCtrl *webhooks.MobileController
	settingsCtrl *settings.MobileController
	eventsCtrl   *events.MobileController

	idGen func() string
}

func newMobileHandler(
	authSvc *auth.Service,
	usersSvc *users.Service,
	devicesSvc *devices.Service,

	messagesCtrl *messages.MobileController,
	webhooksCtrl *webhooks.MobileController,
	settingsCtrl *settings.MobileController,
	eventsCtrl *events.MobileController,

	logger *zap.Logger,
	validator *validator.Validate,
) *mobileHandler {
	const idLength = 21
	idGen, _ := nanoid.Standard(idLength)

	return &mobileHandler{
		Handler: base.Handler{
			Logger:    logger,
			Validator: validator,
		},
		authSvc:    authSvc,
		usersSvc:   usersSvc,
		devicesSvc: devicesSvc,

		messagesCtrl: messagesCtrl,
		webhooksCtrl: webhooksCtrl,
		settingsCtrl: settingsCtrl,
		eventsCtrl:   eventsCtrl,

		idGen: idGen,
	}
}

//	@Summary		Get device information
//	@Description	Returns device information
//	@Tags			Device
//	@Produce		json
//	@Success		200	{object}	smsgateway.MobileDeviceResponse	"Device information"
//	@Failure		500	{object}	smsgateway.ErrorResponse		"Internal server error"
//	@Router			/mobile/v1/device [get]
//
// Get device information.
func (h *mobileHandler) getDevice(device models.Device, c *fiber.Ctx) error {
	res := smsgateway.MobileDeviceResponse{
		ExternalIP: c.IP(),
		Device:     nil,
	}

	if !device.IsEmpty() {
		res.Device = anys.AsPointer(converters.DeviceToDTO(device))
	}

	return c.JSON(res)
}

//	@Summary		Register device
//	@Description	Registers new device for new or existing user. Returns user credentials only for new users
//	@Security		ApiAuth
//	@Security		UserCode
//	@Security		ServerKey
//	@Tags			Device
//	@Accept			json
//	@Produce		json
//	@Param			request	body		smsgateway.MobileRegisterRequest	true	"Device registration request"
//	@Success		201		{object}	smsgateway.MobileRegisterResponse	"Device registered"
//	@Failure		400		{object}	smsgateway.ErrorResponse			"Invalid request"
//	@Failure		401		{object}	smsgateway.ErrorResponse			"Unauthorized (private mode only)"
//	@Failure		429		{object}	smsgateway.ErrorResponse			"Too many requests"
//	@Failure		500		{object}	smsgateway.ErrorResponse			"Internal server error"
//	@Router			/mobile/v1/device [post]
//
// Register device.
func (h *mobileHandler) postDevice(c *fiber.Ctx) error {
	req := new(smsgateway.MobileRegisterRequest)

	if err := h.BodyParserValidator(c, req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var (
		err      error
		userID   string
		password string
	)

	if authUser := userauth.GetUserID(c); authUser != "" {
		userID = authUser
	} else {
		id := h.idGen()
		userID = strings.ToUpper(id[:6])
		password = strings.ToLower(id[7:])

		if _, err = h.usersSvc.Create(userID, password); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	device, err := h.authSvc.RegisterDevice(userID, req.Name, req.PushToken)
	if err != nil {
		return fmt.Errorf("failed to register device: %w", err)
	}

	return c.Status(fiber.StatusCreated).
		JSON(smsgateway.MobileRegisterResponse{
			Id:       device.ID,
			Token:    device.AuthToken,
			Login:    userID,
			Password: password,
		})
}

//	@Summary		Update device
//	@Description	Updates push token for device
//	@Security		MobileToken
//	@Tags			Device
//	@Accept			json
//	@Param			request	body	smsgateway.MobileUpdateRequest	true	"Device update request"
//	@Success		204		"Successfully updated"
//	@Failure		400		{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		403		{object}	smsgateway.ErrorResponse	"Forbidden (wrong device ID)"
//	@Failure		500		{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/mobile/v1/device [patch]
//
// Update device.
func (h *mobileHandler) patchDevice(device models.Device, c *fiber.Ctx) error {
	req := new(smsgateway.MobileUpdateRequest)

	if err := h.BodyParserValidator(c, req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.Id != device.ID {
		return fiber.ErrForbidden
	}

	if err := h.devicesSvc.UpdatePushToken(req.Id, req.PushToken); err != nil {
		h.Logger.Error("failed to update device", zap.Error(err), zap.String("device_id", req.Id))
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update device")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

//	@Summary		Get one-time code for device registration
//	@Description	Returns one-time code for device registration
//	@Security		ApiAuth
//	@Tags			Device
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	smsgateway.MobileUserCodeResponse	"User code"
//	@Failure		500	{object}	smsgateway.ErrorResponse			"Internal server error"
//	@Router			/mobile/v1/user/code [get]
//
// Get user code.
func (h *mobileHandler) getUserCode(userID string, c *fiber.Ctx) error {
	code, err := h.authSvc.GenerateUserCode(c.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to generate user code", zap.Error(err), zap.String("user_id", userID))
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate user code")
	}

	return c.JSON(smsgateway.MobileUserCodeResponse{
		Code:       code.Code,
		ValidUntil: code.ValidUntil,
	})
}

//	@Summary		Change password
//	@Description	Changes the user's password
//	@Security		MobileToken
//	@Tags			Device
//	@Accept			json
//	@Produce		json
//	@Param			request	body		smsgateway.MobileChangePasswordRequest	true	"Password change request"
//	@Success		204		{object}	nil										"Password changed successfully"
//	@Failure		400		{object}	smsgateway.ErrorResponse				"Invalid request"
//	@Failure		401		{object}	smsgateway.ErrorResponse				"Unauthorized"
//	@Failure		500		{object}	smsgateway.ErrorResponse				"Internal server error"
//	@Router			/mobile/v1/user/password [patch]
//
// Change password.
func (h *mobileHandler) changePassword(device models.Device, c *fiber.Ctx) error {
	req := new(smsgateway.MobileChangePasswordRequest)

	if err := h.BodyParserValidator(c, req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := h.usersSvc.ChangePassword(c.Context(), device.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		h.Logger.Error("failed to change password", zap.Error(err))
		return fiber.NewError(fiber.StatusUnauthorized, "failed to change password")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *mobileHandler) Register(router fiber.Router) {
	router = router.Group("/mobile/v1")

	router.Post("/device",
		userauth.NewBasic(h.usersSvc),
		userauth.NewCode(h.authSvc),
		keyauth.New(keyauth.Config{
			Next: func(c *fiber.Ctx) bool {
				// Skip server key authorization in the following cases:
				// 1. Public mode is enabled - allowing open registration
				// 2. User is already authenticated - allowing device registration for existing users
				return h.authSvc.IsPublic() || userauth.HasUser(c)
			},
			Validator: func(_ *fiber.Ctx, token string) (bool, error) {
				err := h.authSvc.AuthorizeRegistration(token)
				if err != nil {
					return false, fmt.Errorf("authorization failed: %w", err)
				}

				return true, nil
			},
		}),
		h.postDevice,
	)

	router.Get("/user/code",
		userauth.NewBasic(h.usersSvc),
		userauth.UserRequired(),
		userauth.WithUserID(h.getUserCode),
	)

	router.Use(
		deviceauth.New(h.authSvc),
	)

	router.Get("/device", deviceauth.WithDevice(h.getDevice))

	router.Use(deviceauth.DeviceRequired())

	router.Patch("/device", deviceauth.WithDevice(h.patchDevice))

	// Should be under `userauth.NewBasic` protection instead of `deviceauth`
	router.Patch("/user/password", deviceauth.WithDevice(h.changePassword))

	h.messagesCtrl.Register(router.Group("/message"))
	h.messagesCtrl.Register(router.Group("/messages"))
	h.webhooksCtrl.Register(router.Group("/webhooks"))
	h.settingsCtrl.Register(router.Group("/settings"))
	h.eventsCtrl.Register(router.Group("/events"))
}
