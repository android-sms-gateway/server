package inbox

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/inbox"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type ThirdPartyController struct {
	base.Handler

	inboxSvc *inbox.Service
}

func NewThirdPartyController(
	inboxSvc *inbox.Service,
	logger *zap.Logger,
	validator *validator.Validate,
) *ThirdPartyController {
	return &ThirdPartyController{
		Handler: base.Handler{
			Logger:    logger,
			Validator: validator,
		},

		inboxSvc: inboxSvc,
	}
}

func (h *ThirdPartyController) Register(router fiber.Router) {
	router.Get("", permissions.RequireScope(ScopeList), userauth.WithUserID(h.list))
	router.Post("/refresh", permissions.RequireScope(ScopeRefresh), userauth.WithUserID(h.refresh))
}

//	@Summary		Get incoming messages
//	@Description	Retrieves incoming messages with filtering and pagination.
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Inbox
//	@Produce		json
//	@Param			type		query		string						false	"Filter incoming messages by type"		Enums(SMS,DATA_SMS,MMS,MMS_DOWNLOADED)
//	@Param			limit		query		int							false	"Maximum number of messages to return"	minimum(1)	maximum(500)	default(50)
//	@Param			offset		query		int							false	"Number of messages to skip"			minimum(0)	default(0)
//	@Param			from		query		string						false	"Start of date range (ISO 8601)"		Format(date-time)
//	@Param			to			query		string						false	"End of date range (ISO 8601)"			Format(date-time)
//	@Param			deviceId	query		string						false	"Device ID"
//	@Success		200			{array}		smsgateway.IncomingMessage	"A list of incoming messages"
//	@Header			200			{integer}	X-Total-Count				"Total number of items available"
//	@Failure		400			{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		401			{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403			{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		500			{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Failure		501			{object}	smsgateway.ErrorResponse	"Not implemented"
//	@Router			/3rdparty/v1/inbox [get]
//
// Get incoming messages.
func (h *ThirdPartyController) list(_ string, _ *fiber.Ctx) error {
	return fiber.NewError(fiber.StatusNotImplemented, "Inbox API is not implemented yet")
}

//	@Summary		Request inbox messages refresh
//	@Description	Refreshes inbox messages. Webhooks are triggered when triggerWebhooks is true.
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Inbox
//	@Accept			json
//	@Produce		json
//	@Param			request	body	smsgateway.InboxRefreshRequest	true	"Refresh inbox request"
//	@Success		202		"Inbox refresh request accepted"
//	@Failure		400		{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		401		{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403		{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		500		{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/3rdparty/v1/inbox/refresh [post]
//
// Request inbox refresh.
func (h *ThirdPartyController) refresh(userID string, c *fiber.Ctx) error {
	req := new(smsgateway.InboxRefreshRequest)
	if err := h.BodyParserValidator(c, req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := h.inboxSvc.Refresh(
		userID,
		&req.DeviceID,
		req.Since,
		req.Until,
		req.MessageTypes,
		&req.TriggerWebhooks,
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusAccepted)
}
