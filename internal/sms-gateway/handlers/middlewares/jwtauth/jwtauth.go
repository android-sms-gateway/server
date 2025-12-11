package jwtauth

import (
	"strings"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/jwt"
	"github.com/gofiber/fiber/v2"
)

func NewJWT(jwtSvc jwt.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")

		if len(token) <= 7 || !strings.EqualFold(token[:7], "Bearer ") {
			return c.Next()
		}

		token = token[7:]

		claims, err := jwtSvc.ParseToken(c.Context(), token)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		userauth.SetUserID(c, claims.UserID)
		permissions.SetScopes(c, claims.Scopes)

		return c.Next()
	}
}
