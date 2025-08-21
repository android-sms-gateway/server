package openapi

import (
	"github.com/android-sms-gateway/server/internal/version"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/swagger"
)

//go:generate swag init --parseDependency --tags=User,System --outputTypes go -d ../../../ -g ./cmd/sms-gateway/main.go -o ../../../internal/sms-gateway/openapi

type Handler struct {
	config Config
}

func New(config Config) *Handler {
	return &Handler{
		config: config,
	}
}

func (s *Handler) Register(router fiber.Router) {
	if !s.config.Enabled {
		return
	}

	SwaggerInfo.Version = version.AppVersion
	SwaggerInfo.Host = s.config.APIHost
	SwaggerInfo.BasePath = s.config.APIPath

	router.Use("*",
		// Pre-middleware: set host/scheme dynamically
		func(c *fiber.Ctx) error {
			if SwaggerInfo.Host == "" {
				SwaggerInfo.Host = c.Hostname()
				SwaggerInfo.BasePath = "/api"
			}

			scheme := "http"
			if c.Secure() {
				scheme = "https"
			}
			SwaggerInfo.Schemes = []string{scheme}
			return c.Next()
		},
		etag.New(etag.Config{Weak: true}),
		swagger.New(swagger.Config{Layout: "BaseLayout"}),
	)
}
