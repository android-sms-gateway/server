package webhooks

import (
	"fmt"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/webhooks"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type thirdPartyControllerParams struct {
	fx.In

	WebhooksSvc *webhooks.Service

	Validator *validator.Validate
	Logger    *zap.Logger
}

type ThirdPartyController struct {
	base.Handler

	webhooksSvc *webhooks.Service
}

func NewThirdPartyController(params thirdPartyControllerParams) *ThirdPartyController {
	return &ThirdPartyController{
		Handler: base.Handler{
			Logger:    params.Logger,
			Validator: params.Validator,
		},
		webhooksSvc: params.WebhooksSvc,
	}
}

//	@Summary		List webhooks
//	@Description	Returns list of registered webhooks
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Webhooks
//	@Produce		json
//	@Success		200	{object}	[]smsgateway.Webhook		"Webhook list"
//	@Failure		401	{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		500	{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/3rdparty/v1/webhooks [get]
//
// List webhooks.
func (h *ThirdPartyController) get(userID string, c *fiber.Ctx) error {
	items, err := h.webhooksSvc.Select(userID)
	if err != nil {
		return fmt.Errorf("failed to select webhooks: %w", err)
	}

	return c.JSON(items)
}

//	@Summary		Register webhook
//	@Description	Registers webhook. If webhook with same ID already exists, it will be replaced
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Webhooks
//	@Accept			json
//	@Produce		json
//	@Param			request	body		smsgateway.Webhook			true	"Webhook"
//	@Success		201		{object}	smsgateway.Webhook			"Created"
//	@Failure		400		{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		401		{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403		{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		500		{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/3rdparty/v1/webhooks [post]
//
// Register webhook.
func (h *ThirdPartyController) post(userID string, c *fiber.Ctx) error {
	dto := new(smsgateway.Webhook)

	if err := h.BodyParserValidator(c, dto); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := h.webhooksSvc.Replace(userID, dto); err != nil {
		if webhooks.IsValidationError(err) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return fmt.Errorf("failed to write webhook: %w", err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto)
}

//	@Summary		Delete webhook
//	@Description	Deletes webhook
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Webhooks
//	@Produce		json
//	@Param			id	path		string						true	"Webhook ID"
//	@Success		204	{object}	object						"Webhook deleted"
//	@Failure		401	{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		500	{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Router			/3rdparty/v1/webhooks/{id} [delete]
//
// Delete webhook.
func (h *ThirdPartyController) delete(userID string, c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.webhooksSvc.Delete(userID, webhooks.WithExtID(id)); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ThirdPartyController) Register(router fiber.Router) {
	router.Get("", permissions.RequireScope(ScopeList), userauth.WithUserID(h.get))
	router.Post("", permissions.RequireScope(ScopeWrite), userauth.WithUserID(h.post))
	router.Delete("/:id", permissions.RequireScope(ScopeDelete), userauth.WithUserID(h.delete))
}
