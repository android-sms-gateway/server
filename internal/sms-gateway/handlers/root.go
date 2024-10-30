package handlers

import (
	"net/http"

	"github.com/capcom6/sms-gateway/pkg/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

type rootHandler struct {
	healthHandler *healthHandler
}

func (h *rootHandler) Register(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/api" {
			return c.Redirect("/api/", fiber.StatusMovedPermanently)
		}

		return c.Next()
	})

	h.healthHandler.Register(app)
	app.Use("/api", filesystem.New(filesystem.Config{
		Root:       http.FS(swagger.Docs),
		PathPrefix: "docs",
		MaxAge:     1 * 24 * 60 * 60,
	}))
	// app.Static("/api", "api")
}

func newRootHandler(healthHandler *healthHandler) *rootHandler {
	return &rootHandler{
		healthHandler: healthHandler,
	}
}
