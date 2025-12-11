package userauth

import (
	"encoding/base64"
	"strings"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/auth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/users"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

const localsUserID = "userID"

// NewBasic returns a middleware that optionally performs HTTP Basic authentication.
// If the "Authorization" header is missing or does not start with "Basic ", the request is passed through unchanged.
// If the header is present, the middleware expects a base64-encoded "username:password" payload, decodes it,
// validates the credentials format, and authenticates the user using the given users service.
// On invalid or failed authentication it returns 401 Unauthorized; on success it stores the user ID in Locals.
func NewBasic(usersSvc *users.Service) fiber.Handler {
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

		user, err := usersSvc.Login(c.Context(), username, password)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		SetUserID(c, user.ID)
		permissions.SetScopes(c, []string{permissions.ScopeAll})

		return c.Next()
	}
}

// NewCode returns a middleware that will check if the request contains a valid
// "Authorization" header in the form of "Code <one-time user authorization code>".
// If the header is valid, the middleware will authorize the user and store the
// user ID in the request's Locals under the key localsUserID. If the header is invalid,
// the middleware will call c.Next() and continue with the request.
func NewCode(authSvc *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get(fiber.HeaderAuthorization)

		if len(auth) <= 5 || !strings.EqualFold(auth[:5], "code ") {
			return c.Next()
		}

		// Get the code
		code := auth[5:]

		user, err := authSvc.AuthorizeUserByCode(c.Context(), code)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		SetUserID(c, user.ID)

		return c.Next()
	}
}

func SetUserID(c *fiber.Ctx, userID string) {
	c.Locals(localsUserID, userID)
}

// HasUser checks if a user is present in the Locals of the given context.
// It returns true if the Locals contain a user ID under the key localsUserID,
// otherwise returns false.
func HasUser(c *fiber.Ctx) bool {
	return GetUserID(c) != ""
}

// GetUserID returns the user ID stored in the Locals of the given context.
// It returns an empty string if the Locals do not contain a user ID under the key localsUserID.
// The user ID is stored in Locals by the NewBasic and NewCode middlewares via SetUserID.
func GetUserID(c *fiber.Ctx) string {
	userID, ok := c.Locals(localsUserID).(string)
	if !ok {
		return ""
	}

	return userID
}

// UserRequired is a middleware that checks if a user is present in the request's Locals.
// If the user is not present, it will return an unauthorized error.
// It is a convenience function that wraps the call to HasUser and calls the
// handler if the user is present.
func UserRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasUser(c) {
			return fiber.ErrUnauthorized
		}

		return c.Next()
	}
}

// WithUserID is a decorator that provides the current user to the handler.
// It assumes that the user is stored in Locals under the key localsUser.
// If the user is not present, it returns 401 Unauthorized.
//
// It is a convenience function that wraps the call to GetUserID and calls the
// handler with the user as the first argument.
func WithUserID(handler func(string, *fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return fiber.ErrUnauthorized
		}

		return handler(userID, c)
	}
}
