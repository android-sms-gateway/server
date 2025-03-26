package userauth

import (
	"encoding/base64"
	"strings"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

const localsUser = "user"

// NewBasic returns a middleware that will check if the request contains a valid
// "Authorization" header in the form of "Basic <base64 encoded username:password>".
// If the header is valid, the middleware will authorize the user and store the
// user in the request's Locals under the key LocalsUser. If the header is invalid,
// the middleware will call c.Next() and continue with the request.
func NewBasic(authSvc *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get(fiber.HeaderAuthorization)

		if len(auth) <= 6 || !strings.EqualFold(auth[:6], "basic ") {
			return c.Next()
		}

		// Decode the header contents
		raw, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			return fiber.ErrUnauthorized
		}

		// Get the credentials
		creds := utils.UnsafeString(raw)

		// Check if the credentials are in the correct form
		// which is "username:password".
		index := strings.Index(creds, ":")
		if index == -1 {
			return fiber.ErrUnauthorized
		}

		// Get the username and password
		username := creds[:index]
		password := creds[index+1:]

		user, err := authSvc.AuthorizeUser(username, password)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		c.Locals(localsUser, user)

		return c.Next()
	}
}

// NewCode returns a middleware that will check if the request contains a valid
// "Authorization" header in the form of "Code <one-time user authorization code>".
// If the header is valid, the middleware will authorize the user and store the
// user in the request's Locals under the key LocalsUser. If the header is invalid,
// the middleware will call c.Next() and continue with the request.
func NewCode(authSvc *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get(fiber.HeaderAuthorization)

		if len(auth) <= 5 || !strings.EqualFold(auth[:5], "code ") {
			return c.Next()
		}

		// Get the code
		code := auth[5:]

		user, err := authSvc.AuthorizeUserByCode(code)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		c.Locals(localsUser, user)

		return c.Next()
	}
}

// HasUser checks if a user is present in the Locals of the given context.
// It returns true if the Locals contain a user under the key LocalsUser,
// otherwise returns false.
func HasUser(c *fiber.Ctx) bool {
	return c.Locals(localsUser) != nil
}

// GetUser returns the user stored in the Locals under the key LocalsUser.
// It is a convenience function that wraps the call to c.Locals(LocalsUser) and
// casts the result to models.User.
//
// It panics if the value stored in Locals is not a models.User.
func GetUser(c *fiber.Ctx) models.User {
	return c.Locals(localsUser).(models.User)
}

// UserRequired is a middleware that ensures a user is present in the request's Locals.
// If a user is not found, it returns an unauthorized error, otherwise it passes control
// to the next handler in the stack.
func UserRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasUser(c) {
			return fiber.ErrUnauthorized
		}

		return c.Next()
	}
}

// WithUser is a decorator that provides the current user to the handler.
// It assumes that the user is stored in the Locals under the key LocalsUser.
// If the user is not present, it will panic.
//
// It is a convenience function that wraps the call to GetUser and calls the
// handler with the user as the first argument.
func WithUser(handler func(models.User, *fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return handler(GetUser(c), c)
	}
}
