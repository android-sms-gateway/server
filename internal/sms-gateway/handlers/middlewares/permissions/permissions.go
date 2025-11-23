package permissions

import (
	"slices"

	"github.com/gofiber/fiber/v2"
)

const (
	ScopeAll = "all:any"

	localsScopes = "user:scopes"
)

func SetScopes(c *fiber.Ctx, scopes []string) {
	c.Locals(localsScopes, scopes)
}

func HasScope(c *fiber.Ctx, scope string) bool {
	scopes, ok := c.Locals(localsScopes).([]string)
	if !ok {
		return false
	}

	return slices.ContainsFunc(scopes, func(item string) bool { return item == scope || item == ScopeAll })
}

func RequireScope(scope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasScope(c, scope) {
			return fiber.NewError(fiber.StatusForbidden, "scope required: "+scope)
		}

		return c.Next()
	}
}
