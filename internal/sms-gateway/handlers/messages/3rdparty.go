package messages

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/converters"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/capcom6/go-helpers/slices"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	route3rdPartyGetMessage = "3rdparty.get.message"
)

type thirdPartyControllerParams struct {
	fx.In

	MessagesSvc *messages.Service
	DevicesSvc  *devices.Service

	Validator *validator.Validate
	Logger    *zap.Logger
}

type ThirdPartyController struct {
	base.Handler

	messagesSvc *messages.Service
	devicesSvc  *devices.Service
}

func NewThirdPartyController(params thirdPartyControllerParams) *ThirdPartyController {
	return &ThirdPartyController{
		Handler: base.Handler{
			Logger:    params.Logger,
			Validator: params.Validator,
		},
		messagesSvc: params.MessagesSvc,
		devicesSvc:  params.DevicesSvc,
	}
}

//	@Summary		Enqueue message
//	@Description	Enqueues a message for sending. If `deviceId` is set, the specified device is used; otherwise a random registered device is chosen.
//	@Security		ApiAuth
//	@Tags			User, Messages
//	@Accept			json
//	@Produce		json
//	@Param			skipPhoneValidation	query		bool							false	"Skip phone validation"
//	@Param			deviceActiveWithin	query		int								false	"Filter devices active within the specified number of hours"	default(0)	minimum(0)
//	@Param			request				body		smsgateway.Message				true	"Send message request"
//	@Success		202					{object}	smsgateway.GetMessageResponse	"Message enqueued"
//	@Failure		400					{object}	smsgateway.ErrorResponse		"Invalid request"
//	@Failure		401					{object}	smsgateway.ErrorResponse		"Unauthorized"
//	@Failure		409					{object}	smsgateway.ErrorResponse		"Message with such ID already exists"
//	@Failure		500					{object}	smsgateway.ErrorResponse		"Internal server error"
//	@Header			202					{string}	Location						"Get message state URL"
//	@Router			/3rdparty/v1/messages [post]
//
// Enqueue message.
func (h *ThirdPartyController) post(user models.User, c *fiber.Ctx) error {
	var params thirdPartyPostQueryParams
	if err := h.QueryParserValidator(c, &params); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var req smsgateway.Message
	if err := h.BodyParserValidator(c, &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	device, err := h.devicesSvc.GetAny(
		user.ID,
		req.DeviceID,
		time.Duration(params.DeviceActiveWithin)*time.Hour,
	)
	if err != nil {
		h.Logger.Error(
			"failed to select device",
			zap.Error(err),
			zap.String("user_id", user.ID),
			zap.String("device_id", req.DeviceID),
		)

		return fmt.Errorf("failed to select device: %w", err)
	}

	var textContent *messages.TextMessageContent
	var dataContent *messages.DataMessageContent
	if text := req.GetTextMessage(); text != nil {
		textContent = &messages.TextMessageContent{
			Text: text.Text,
		}
	} else if data := req.GetDataMessage(); data != nil {
		dataContent = &messages.DataMessageContent{
			Data: data.Data,
			Port: data.Port,
		}
	} else {
		return fiber.NewError(fiber.StatusBadRequest, "No message content provided")
	}

	msg := messages.MessageIn{
		ID: req.ID,

		TextContent: textContent,
		DataContent: dataContent,

		PhoneNumbers: req.PhoneNumbers,
		IsEncrypted:  req.IsEncrypted,

		SimNumber:          req.SimNumber,
		WithDeliveryReport: req.WithDeliveryReport,
		TTL:                req.TTL,
		ValidUntil:         req.ValidUntil,
		Priority:           req.Priority,
	}
	state, err := h.messagesSvc.Enqueue(
		*device,
		msg,
		messages.EnqueueOptions{SkipPhoneValidation: params.SkipPhoneValidation},
	)
	if err != nil {
		h.Logger.Error(
			"failed to enqueue message",
			zap.Error(err),
			zap.String("user_id", user.ID),
			zap.String("device_id", req.DeviceID),
		)

		return fmt.Errorf("failed to enqueue message: %w", err)
	}

	location, err := c.GetRouteURL(route3rdPartyGetMessage, fiber.Map{
		"id": state.ID,
	})
	if err != nil {
		h.Logger.Warn(
			"failed to get route URL",
			zap.String("route", route3rdPartyGetMessage),
			zap.String("id", state.ID),
			zap.Error(err),
		)
	} else {
		c.Location(location)
	}

	return c.Status(fiber.StatusAccepted).
		JSON(smsgateway.GetMessageResponse{
			ID:          state.ID,
			DeviceID:    state.DeviceID,
			State:       smsgateway.ProcessingState(state.State),
			IsHashed:    state.IsHashed,
			IsEncrypted: state.IsEncrypted,
			Recipients:  state.Recipients,
			States:      state.States,
		})
}

//	@Summary		Get messages
//	@Description	Retrieves a list of messages with filtering and pagination
//	@Security		ApiAuth
//	@Tags			User, Messages
//	@Produce		json
//	@Param			from		query		string							false	"Start date in RFC3339 format"			Format(date-time)
//	@Param			to			query		string							false	"End date in RFC3339 format"			Format(date-time)
//	@Param			state		query		string							false	"Filter messages by processing state"	Enum(Pending, Processed, Sent, Delivered, Failed)
//	@Param			deviceId	query		string							false	"Filter by device ID"					min(21)		max(21)
//	@Param			limit		query		int								false	"Pagination limit"						default(50)	min(1)	max(100)
//	@Param			offset		query		int								false	"Pagination offset"						default(0)
//	@Success		200			{object}	smsgateway.GetMessagesResponse	"A list of messages"
//	@Failure		400			{object}	smsgateway.ErrorResponse		"Invalid request"
//	@Failure		401			{object}	smsgateway.ErrorResponse		"Unauthorized"
//	@Failure		500			{object}	smsgateway.ErrorResponse		"Internal server error"
//	@Router			/3rdparty/v1/messages [get]
//
// Get message history.
func (h *ThirdPartyController) list(user models.User, c *fiber.Ctx) error {
	params := new(thirdPartyGetQueryParams)
	if err := h.QueryParserValidator(c, params); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	messages, total, err := h.messagesSvc.SelectStates(user, params.ToFilter(), params.ToOptions())
	if err != nil {
		h.Logger.Error("failed to get message history", zap.Error(err), zap.String("user_id", user.ID))
		return fiber.NewError(fiber.StatusInternalServerError, "failed to retrieve message history")
	}

	c.Set("X-Total-Count", strconv.Itoa(int(total)))
	return c.JSON(
		slices.Map(messages, converters.MessageStateToDTO),
	)
}

//	@Summary		Get message state
//	@Description	Returns message state by ID
//	@Security		ApiAuth
//	@Tags			User, Messages
//	@Produce		json
//	@Param			id	path		string							true	"Message ID"
//	@Success		200	{object}	smsgateway.GetMessageResponse	"Message state"
//	@Failure		400	{object}	smsgateway.ErrorResponse		"Invalid request"
//	@Failure		401	{object}	smsgateway.ErrorResponse		"Unauthorized"
//	@Failure		500	{object}	smsgateway.ErrorResponse		"Internal server error"
//	@Router			/3rdparty/v1/messages/{id} [get]
//
// Get message state.
func (h *ThirdPartyController) get(user models.User, c *fiber.Ctx) error {
	id := c.Params("id")

	state, err := h.messagesSvc.GetState(user, id)
	if err != nil {
		if errors.Is(err, messages.ErrMessageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}

		h.Logger.Error("failed to get message state", zap.Error(err), zap.String("user_id", user.ID))
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get message state")
	}

	return c.JSON(converters.MessageStateToDTO(*state))
}

//	@Summary		Request inbox messages export
//	@Description	Initiates process of inbox messages export via webhooks. For each message the `sms:received` webhook will be triggered. The webhooks will be triggered without specific order.
//	@Security		ApiAuth
//	@Tags			User, Messages
//	@Accept			json
//	@Produce		json
//	@Param			request	body		smsgateway.MessagesExportRequest	true	"Export inbox request"
//	@Success		202		{object}	object								"Inbox export request accepted"
//	@Failure		400		{object}	smsgateway.ErrorResponse			"Invalid request"
//	@Failure		401		{object}	smsgateway.ErrorResponse			"Unauthorized"
//	@Failure		500		{object}	smsgateway.ErrorResponse			"Internal server error"
//	@Router			/3rdparty/v1/messages/inbox/export [post]
//
// Export inbox.
func (h *ThirdPartyController) postInboxExport(user models.User, c *fiber.Ctx) error {
	req := new(smsgateway.MessagesExportRequest)
	if err := h.BodyParserValidator(c, req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	device, err := h.devicesSvc.Get(user.ID, devices.WithID(req.DeviceID))
	if err != nil {
		if errors.Is(err, devices.ErrNotFound) {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid device ID")
		}

		h.Logger.Error("failed to get device", zap.Error(err), zap.String("user_id", user.ID))
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get device")
	}

	if expErr := h.messagesSvc.ExportInbox(device, req.Since, req.Until); expErr != nil {
		h.Logger.Error("failed to export inbox", zap.Error(expErr), zap.String("user_id", user.ID))
		return fiber.NewError(fiber.StatusInternalServerError, "failed to export inbox")
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func (h *ThirdPartyController) errorHandler(c *fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}

	var fiberError *fiber.Error
	if errors.As(err, &fiberError) {
		return fiberError
	}

	var msgValidationError messages.ValidationError
	switch {
	case errors.As(err, &msgValidationError):
		fallthrough
	case errors.Is(err, messages.ErrMultipleMessagesFound):
		fallthrough
	case errors.Is(err, messages.ErrNoContent):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())

	case errors.Is(err, messages.ErrMessageNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())

	case errors.Is(err, messages.ErrMessageAlreadyExists):
		return fiber.NewError(fiber.StatusConflict, err.Error())

	case errors.Is(err, devices.ErrNotFound):
		fallthrough
	case errors.Is(err, devices.ErrInvalidFilter):
		fallthrough
	case errors.Is(err, devices.ErrInvalidUser):
		fallthrough
	case errors.Is(err, devices.ErrMoreThanOne):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	h.Logger.Error("failed to handle request", zap.Error(err))
	return fiber.NewError(fiber.StatusInternalServerError, "failed to handle request")
}

func (h *ThirdPartyController) Register(router fiber.Router) {
	router.Use(h.errorHandler)

	router.Get("", userauth.WithUser(h.list))
	router.Post("", userauth.WithUser(h.post))
	router.Get(":id", userauth.WithUser(h.get)).Name(route3rdPartyGetMessage)

	router.Post("inbox/export", userauth.WithUser(h.postInboxExport))
}
