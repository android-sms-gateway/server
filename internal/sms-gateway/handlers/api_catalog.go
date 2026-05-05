package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.uber.org/zap"
)

type APICatalogHandler struct {
	config Config
	logger *zap.Logger
}

func newAPICatalogHandler(cfg Config, logger *zap.Logger) *APICatalogHandler {
	return &APICatalogHandler{
		config: cfg,
		logger: logger.Named("api_catalog"),
	}
}

func (h *APICatalogHandler) get(c *fiber.Ctx) error {
	const (
		fieldHref = "href"
		fieldType = "type"
	)

	c.Set(fiber.HeaderContentType, `application/linkset+json; profile="https://www.rfc-editor.org/info/rfc9727"`)
	c.Set(fiber.HeaderLink, `</.well-known/api-catalog>; rel="api-catalog"`)
	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")

	host := h.getHost(c)
	path := h.getPath(c)
	linkset := fiber.Map{
		"linkset": []fiber.Map{
			{
				"anchor": fmt.Sprintf("https://%s/%s/3rdparty/v1", host, path),
				"service-desc": []fiber.Map{
					{
						fieldHref: fmt.Sprintf("https://%s/%s/docs/doc.json", host, path),
						fieldType: "application/json",
					},
				},
				"service-doc": []fiber.Map{
					{
						fieldHref: "https://docs.sms-gate.app/",
						fieldType: "text/html",
					},
				},
				"status": []fiber.Map{
					{
						fieldHref: fmt.Sprintf("https://%s/%s/3rdparty/v1/health/ready", host, path),
						fieldType: "application/json",
					},
				},
			},
		},
	}

	return c.JSON(linkset)
}

func (h *APICatalogHandler) head(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, `application/linkset+json; profile="https://www.rfc-editor.org/info/rfc9727"`)
	c.Set("Link", `</.well-known/api-catalog>; rel="api-catalog"`)
	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.SendStatus(fiber.StatusOK)
}

func (h *APICatalogHandler) getHost(c *fiber.Ctx) string {
	if h.config.PublicHost != "" {
		return h.config.PublicHost
	}
	return c.Hostname()
}

func (h *APICatalogHandler) getPath(_ *fiber.Ctx) any {
	return strings.TrimLeft(h.config.PublicPath, "/")
}

func (h *APICatalogHandler) Register(app *fiber.App) {
	const limit = 60

	rateLimiter := limiter.New(limiter.Config{
		Max:               limit,
		Expiration:        time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
	})

	group := app.Group("/.well-known/api-catalog", rateLimiter)
	group.Get("", h.get)
	group.Head("", h.head)
}
