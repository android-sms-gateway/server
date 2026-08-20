package inbox

import (
	"errors"

	"github.com/android-sms-gateway/server/internal/sms-gateway/inbox"
	"github.com/gofiber/fiber/v2"
)

func errorHandler(c *fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, inbox.ErrNotEncrypted):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, inbox.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, inbox.ErrEmptyBatch):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return err //nolint:wrapcheck // passed through to fiber's error handler
}
