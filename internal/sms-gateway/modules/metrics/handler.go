//nolint:revive // ignore package name
package metrics

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
)

type HTTPHandler struct {
}

func (h *HTTPHandler) Register(app *fiber.App) {
	promhandler := fiberprometheus.New("")
	promhandler.RegisterAt(app, "/metrics")

	app.Use(promhandler.Middleware)
}

func newHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}
