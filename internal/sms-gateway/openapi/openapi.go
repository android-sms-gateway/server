package openapi

import (
	"github.com/android-sms-gateway/server/internal/version"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/swagger"
)

//go:generate swag init --parseDependency --tags=User,System --outputTypes go -d ../../../ -g ./cmd/sms-gateway/main.go -o ../../../internal/sms-gateway/openapi

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (s *Handler) Register(router fiber.Router, publicHost, publicPath string) {
	SwaggerInfo.Version = version.AppVersion
	SwaggerInfo.Host = publicHost
	SwaggerInfo.BasePath = publicPath

	router.Use("*",
		// Pre-middleware: set host/scheme dynamically
		func(c *fiber.Ctx) error {
			if SwaggerInfo.Host == "" {
				SwaggerInfo.Host = c.Hostname()
			}

			SwaggerInfo.Schemes = []string{c.Protocol()}
			return c.Next()
		},
		etag.New(etag.Config{Weak: true}),
		swagger.New(swagger.Config{Layout: "BaseLayout", URL: "doc.json"}),
	)
}
