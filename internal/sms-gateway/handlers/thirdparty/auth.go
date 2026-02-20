package thirdparty

import (
	"errors"
	"fmt"
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthHandler struct {
	base.Handler

	jwtSvc jwt.Service
}

func NewAuthHandler(
	jwtSvc jwt.Service,

	logger *zap.Logger,
	validator *validator.Validate,
) *AuthHandler {
	return &AuthHandler{
		Handler: base.Handler{Logger: logger, Validator: validator},

		jwtSvc: jwtSvc,
	}
}

func (h *AuthHandler) Register(router fiber.Router) {
	router.Use(h.errorHandler)
	router.Post("/token", permissions.RequireScope(ScopeTokensManage), userauth.WithUserID(h.postToken))
	router.Delete("/token/:jti", permissions.RequireScope(ScopeTokensManage), userauth.WithUserID(h.deleteToken))
}

//	@Summary		Generate token
//	@Description	Generate new access token with specified scopes and ttl\r\n\r\n*Not available in Local Server mode*
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		smsgateway.TokenRequest		true	"Request"
//	@Success		201		{object}	smsgateway.TokenResponse	"Token"
//	@Failure		400		{object}	smsgateway.ErrorResponse	"Invalid request"
//	@Failure		401		{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403		{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		500		{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Failure		501		{object}	smsgateway.ErrorResponse	"Not implemented"
//	@Router			/3rdparty/v1/auth/token [post]
//
// Generate token.
func (h *AuthHandler) postToken(userID string, c *fiber.Ctx) error {
	req := new(smsgateway.TokenRequest)
	if err := h.BodyParserValidator(c, req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	token, err := h.jwtSvc.GenerateToken(
		c.Context(),
		userID,
		req.Scopes,
		time.Duration(req.TTL)*time.Second, //nolint:gosec // validated in the service
	)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	return c.Status(fiber.StatusCreated).JSON(smsgateway.TokenResponse{
		ID:          token.ID,
		TokenType:   "Bearer",
		AccessToken: token.AccessToken,
		ExpiresAt:   token.ExpiresAt,
	})
}

//	@Summary		Revoke token
//	@Description	Revoke access token with specified jti
//	@Security		ApiAuth
//	@Security		JWTAuth
//	@Tags			User, Auth
//	@Produce		json
//	@Param			jti	path	string	true	"JWT ID"
//	@Success		204	"No Content"
//	@Failure		401	{object}	smsgateway.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	smsgateway.ErrorResponse	"Forbidden"
//	@Failure		500	{object}	smsgateway.ErrorResponse	"Internal server error"
//	@Failure		501	{object}	smsgateway.ErrorResponse	"Not implemented"
//	@Router			/3rdparty/v1/auth/token/{jti} [delete]
//
// Revoke token.
func (h *AuthHandler) deleteToken(userID string, c *fiber.Ctx) error {
	jti := c.Params("jti")

	if err := h.jwtSvc.RevokeToken(c.Context(), userID, jti); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) errorHandler(c *fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, jwt.ErrInvalidParams):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())

	case errors.Is(err, jwt.ErrInitFailed):
		fallthrough
	case errors.Is(err, jwt.ErrInvalidConfig):
		return fiber.NewError(
			fiber.StatusInternalServerError,
			"token service not configured, contact your administrator",
		)

	case errors.Is(err, jwt.ErrDisabled):
		return fiber.NewError(fiber.StatusNotImplemented, "token service disabled, contact your administrator")
	}

	return err //nolint:wrapcheck // passed through to fiber's error handler
}
