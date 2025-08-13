package messages

import (
	"errors"
	"fmt"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/converters"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/capcom6/go-helpers/slices"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type mobileControllerParams struct {
	fx.In

	MessagesSvc *messages.Service

	Validator *validator.Validate
	Logger    *zap.Logger
}

type MobileController struct {
	base.Handler

	messagesSvc *messages.Service
}

//	@Summary		Get messages for sending
//	@Description	Returns list of pending messages
//	@Security		MobileToken
//	@Tags			Device, Messages
//	@Accept			json
//	@Produce		json
//	@Param			order	query		string									false	"Message processing order: lifo (default) or fifo"	Enums(lifo,fifo) default(lifo)
//	@Success		200		{object}	smsgateway.MobileGetMessagesResponse	"List of pending messages"
//	@Failure		400		{object}	smsgateway.ErrorResponse				"Invalid request"
//	@Failure		500		{object}	smsgateway.ErrorResponse				"Internal server error"
//	@Router			/mobile/v1/message [get]
//
// Get messages for sending
func (h *MobileController) list(device models.Device, c *fiber.Ctx) error {
	// Get and validate order parameter
	params := mobileGetQueryParams{}
	if err := h.QueryParserValidator(c, &params); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	msgs, err := h.messagesSvc.SelectPending(device.ID, params.OrderOrDefault())
	if err != nil {
		return fmt.Errorf("can't get messages: %w", err)
	}

	return c.JSON(
		smsgateway.MobileGetMessagesResponse(
			slices.Map(
				msgs,
				converters.MessageToMobileDTO,
			),
		),
	)
}

//	@Summary		Update message state
//	@Description	Updates message state
//	@Security		MobileToken
//	@Tags			Device, Messages
//	@Accept			json
//	@Produce		json
//	@Param			request	body		smsgateway.MobilePatchMessageRequest	true	"List of message state updates"
//	@Success		204		{object}	nil										"Successfully updated"
//	@Failure		400		{object}	smsgateway.ErrorResponse				"Invalid request"
//	@Failure		500		{object}	smsgateway.ErrorResponse				"Internal server error"
//	@Router			/mobile/v1/message [patch]
//
// Update message state
func (h *MobileController) patch(device models.Device, c *fiber.Ctx) error {
	var req smsgateway.MobilePatchMessageRequest
	if err := h.BodyParserValidator(c, &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	for _, v := range req {
		messageState := messages.MessageStateIn{
			ID:         v.ID,
			State:      messages.ProcessingState(v.State),
			Recipients: v.Recipients,
			States:     v.States,
		}

		err := h.messagesSvc.UpdateState(device.ID, messageState)
		if err != nil && !errors.Is(err, messages.ErrMessageNotFound) {
			h.Logger.Error("Can't update message status",
				zap.String("message_id", v.ID),
				zap.Error(err),
			)
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MobileController) Register(router fiber.Router) {
	router.Get("", deviceauth.WithDevice(h.list))
	router.Patch("", deviceauth.WithDevice(h.patch))
}

func NewMobileController(params mobileControllerParams) *MobileController {
	return &MobileController{
		Handler: base.Handler{
			Logger:    params.Logger.Named("messages"),
			Validator: params.Validator,
		},
		messagesSvc: params.MessagesSvc,
	}
}
