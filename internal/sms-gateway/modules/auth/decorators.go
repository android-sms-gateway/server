package auth

import (
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/gofiber/fiber/v2"
)

func WithDevice(handler func(models.Device, *fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return handler(c.Locals("device").(models.Device), c)
	}
}
