package handlers

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/openapi"
	"github.com/gofiber/fiber/v2"
)

type rootHandler struct {
	healthHandler  *healthHandler
	openapiHandler *openapi.Handler
}

func (h *rootHandler) Register(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/api" || c.Path() == "/api/" {
			return c.Redirect("/api/docs/", fiber.StatusMovedPermanently)
		}

		return c.Next()
	})

	h.healthHandler.Register(app)
	h.openapiHandler.Register(app.Group("/api/docs"))
}

func newRootHandler(healthHandler *healthHandler, openapiHandler *openapi.Handler) *rootHandler {
	return &rootHandler{
		healthHandler:  healthHandler,
		openapiHandler: openapiHandler,
	}
}
