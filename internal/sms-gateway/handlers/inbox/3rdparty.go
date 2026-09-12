package inbox

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/inbox"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
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
	router.Use(errorHandler)
	router.Get("", permissions.RequireScope(ScopeList), userauth.WithUserID(h.list))
	router.Get("/:id/attachments/:partId", permissions.RequireScope(ScopeRead), userauth.WithUserID(h.getAttachment))
	router.Post("/refresh", permissions.RequireScope(ScopeRefresh), userauth.WithUserID(h.refresh))
}

//	@Summary		Get inbox messages
//	@Description	Retrieves inbox messages with filtering and pagination.
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Inbox
//	@Produce		json
//	@Param			type		query		string						false	"Filter inbox messages by type"			Enums(SMS,DATA_SMS,MMS,MMS_DOWNLOADED)
//	@Param			limit		query		int							false	"Maximum number of messages to return"	minimum(1)	maximum(500)	default(50)
//	@Param			offset		query		int							false	"Number of messages to skip"			minimum(0)	default(0)
//	@Param			from		query		string						false	"Start of date range (ISO 8601)"		Format(date-time)
//	@Param			to			query		string						false	"End of date range (ISO 8601)"			Format(date-time)
//	@Param			deviceId	query		string						false	"Device ID"
//	@Success		200			{array}		smsgateway.IncomingMessage	"A list of inbox messages"
//	@Header			200			{integer}	X-Total-Count				"Total number of items available"
//	@Failure		400			{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		401			{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403			{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		500			{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Failure		501			{object}	smsgateway.ErrorResponse	"Not implemented"
//	@Router			/3rdparty/v1/inbox [get]
//
// Get inbox messages.
func (h *ThirdPartyController) list(userID string, c *fiber.Ctx) error {
	var params thirdPartyListParams
	if err := c.QueryParser(&params); err != nil {
		h.Logger.Error("failed to parse query parameters", zap.Error(err))
		return fiber.NewError(fiber.StatusBadRequest, "failed to parse query parameters")
	}

	messages, total, err := h.inboxSvc.List(userID, params.toFilter(), params.toOptions())
	if err != nil {
		h.Logger.Error("failed to list inbox messages", zap.Error(err), zap.String("user_id", userID))
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list inbox messages")
	}

	result := make([]smsgateway.IncomingMessage, len(messages))
	for i, m := range messages {
		atts := make([]smsgateway.InboxAttachment, len(m.Attachments))
		for j, a := range m.Attachments {
			atts[j] = smsgateway.InboxAttachment{
				PartID:      a.PartID,
				Name:        a.Name,
				Size:        lo.FromPtrOr(a.Size, 0),
				ContentType: a.ContentType,
			}
		}
		result[i] = smsgateway.IncomingMessage{
			ID:             m.ExtID,
			Type:           smsgateway.IncomingMessageType(m.Type),
			Sender:         m.Sender,
			Recipient:      m.Recipient,
			SimNumber:      m.SimNumber,
			ContentPreview: m.Content,
			IsEncrypted:    m.IsEncrypted,
			CreatedAt:      m.CreatedAt,
			Attachments:    atts,
		}
	}

	c.Set("X-Total-Count", strconv.Itoa(int(total)))
	return c.JSON(result)
}

//	@Summary		Request inbox messages refresh
//	@Description	Refreshes inbox messages. Webhook triggering depends on the `webhookDelivery` parameter.
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
		req.DeviceID,
		req.Since,
		req.Until,
		req.MessageTypes,
		req.ResolveWebhookDelivery(),
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusAccepted)
}

//	@Summary		Get attachment
//	@Description	Downloads an attachment from an inbox message by message ID and part ID.
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Inbox
//	@Produce		application/octet-stream
//	@Param			id		path		string						true	"Inbox message ID"
//	@Param			partId	path		int							true	"Attachment part ID"
//	@Success		200		{file}		binary						"Attachment file"
//	@Failure		400		{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		401		{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403		{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		404		{object}	smsgateway.ErrorResponse	"Not found"
//	@Failure		500		{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/3rdparty/v1/inbox/{id}/attachments/{partId} [get]
//
// Get attachment.
func (h *ThirdPartyController) getAttachment(userID string, c *fiber.Ctx) error {
	id := c.Params("id")
	partID, err := strconv.ParseInt(c.Params("partId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid partId")
	}

	att, err := h.inboxSvc.GetAttachment(c.Context(), userID, id, partID)
	if err != nil {
		if errors.Is(err, inbox.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "attachment not found")
		}
		h.Logger.Error("failed to get attachment", zap.Error(err), zap.String("user_id", userID))
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get attachment")
	}

	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Name))
	c.Set("Content-Type", att.ContentType)

	return c.Send(att.Data)
}
