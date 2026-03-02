package permissions

import (
	"slices"

	"github.com/gofiber/fiber/v2"
)

type contextKey string

const (
	ScopeAll = "all:any"

	localsScopes = contextKey("scopes")
)

func SetScopes(c *fiber.Ctx, scopes []string) {
	c.Locals(localsScopes, scopes)
}

func HasScope(c *fiber.Ctx, scope string, opts *options) bool {
	scopes, ok := c.Locals(localsScopes).([]string)
	if !ok {
		return false
	}

	return slices.ContainsFunc(
		scopes,
		func(item string) bool { return item == scope || (!opts.exact && item == ScopeAll) },
	)
}

func RequireScope(scope string, opts ...Option) fiber.Handler {
	o := &options{
		exact: false,
	}
	for _, opt := range opts {
		opt(o)
	}

	return func(c *fiber.Ctx) error {
		if !HasScope(c, scope, o) {
			return fiber.NewError(fiber.StatusForbidden, "scope required: "+scope)
		}

		return c.Next()
	}
}
