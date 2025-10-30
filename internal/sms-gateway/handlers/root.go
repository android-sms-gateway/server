package handlers

import (
	"path"
	"strings"

	"github.com/android-sms-gateway/server/internal/sms-gateway/openapi"
	"github.com/gofiber/fiber/v2"
)

type rootHandler struct {
	config Config

	healthHandler  *HealthHandler
	openapiHandler *openapi.Handler
}

func (h *rootHandler) Register(app *fiber.App) {
	if h.config.PublicPath != "/api" {
		app.Use(func(c *fiber.Ctx) error {
			err := c.Next()

			location := c.GetRespHeader(fiber.HeaderLocation)
			if after, ok := strings.CutPrefix(location, "/api"); ok {
				c.Set(fiber.HeaderLocation, path.Join(h.config.PublicPath, after))
			}

			return err
		})
	}

	h.healthHandler.Register(app)

	h.registerOpenAPI(app)
}

func (h *rootHandler) registerOpenAPI(router fiber.Router) {
	if !h.config.OpenAPIEnabled {
		return
	}

	router.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/api" || c.Path() == "/api/" {
			return c.Redirect("/api/docs", fiber.StatusMovedPermanently)
		}

		return c.Next()
	})
	h.openapiHandler.Register(router.Group("/api/docs"), h.config.PublicHost, h.config.PublicPath)
}

func newRootHandler(cfg Config, healthHandler *HealthHandler, openapiHandler *openapi.Handler) *rootHandler {
	return &rootHandler{
		config: cfg,

		healthHandler:  healthHandler,
		openapiHandler: openapiHandler,
	}
}
