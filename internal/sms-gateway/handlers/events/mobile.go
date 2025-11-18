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

//	@Summary		Get events
//	@Description	Returns events stream for a device
//	@Security		MobileToken
//	@Tags			Device, Events
//	@x-sse			true
//	@Produce		text/event-stream
//	@Header			200	{string}	Content-Type				"text/event-stream"
//	@Header			200	{string}	Transfer-Encoding			"chunked"
//	@Header			200	{string}	Connection					"keep-alive"
//	@Header			200	{string}	Cache-Control				"no-cache"
//	@Success		200	{string}	string						"Event"
//	@Failure		401	{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		500	{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/mobile/v1/events [get]
//
// Get events.
func (h *MobileController) get(device models.Device, c *fiber.Ctx) error {
	return h.sseSvc.Handler(device.ID, c) //nolint:wrapcheck //wrapped internally
}

func (h *MobileController) Register(router fiber.Router) {
	router.Get("", deviceauth.WithDevice(h.get))
}
