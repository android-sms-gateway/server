package jwtauth

import (
	"errors"
	"strings"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/jwt"
	"github.com/android-sms-gateway/server/internal/sms-gateway/users"
	"github.com/gofiber/fiber/v2"
)

func NewJWT(jwtSvc jwt.Service, usersSvc *users.Service) fiber.Handler {
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

		user, err := usersSvc.GetByUsername(claims.UserID)
		if err != nil {
			if !errors.Is(err, users.ErrNotFound) {
				return fiber.ErrInternalServerError
			}
			return fiber.ErrUnauthorized
		}

		userauth.SetUser(c, user)
		permissions.SetScopes(c, claims.Scopes)

		return c.Next()
	}
}
