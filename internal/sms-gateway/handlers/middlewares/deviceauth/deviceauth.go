package deviceauth

import (
	"errors"
	"strings"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/auth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/gofiber/fiber/v2"
)

const LocalsDevice = "device"

// New returns a middleware that will check if the request contains a valid
// "Authorization" header in the form of "Bearer <device auth token>".
// If the header is valid, the middleware will authorize the device and store the
// device in the request's Locals under the key LocalsDevice. If the header is
// invalid, the middleware will call c.Next() and continue with the request.
func New(authSvc *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		auth := c.Get(fiber.HeaderAuthorization)

		if len(auth) <= 7 || !strings.EqualFold(auth[:7], "Bearer ") {
			return c.Next()
		}

		// Get the token
		token := auth[7:]

		device, err := authSvc.AuthorizeDevice(token)
		if errors.Is(err, devices.ErrNotFound) {
			return c.Next()
		}
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		c.Locals(LocalsDevice, device)

		return c.Next()
	}
}

// HasDevice checks if a device is present in the Locals of the given context.
// It returns true if the Locals contain a device under the key LocalsDevice,
// otherwise returns false.
func HasDevice(c *fiber.Ctx) bool {
	return c.Locals(LocalsDevice) != nil
}

// GetDevice returns the device stored in the Locals under the key LocalsDevice.
// If the Locals do not contain a device, it returns an empty device.
func GetDevice(c *fiber.Ctx) models.Device {
	device, ok := c.Locals(LocalsDevice).(models.Device)
	if !ok {
		return models.Device{}
	}

	return device
}

// DeviceRequired is a middleware that ensures a device is present in the request's Locals.
// If a device is not found, it returns an unauthorized error, otherwise it passes control
// to the next handler in the stack.
func DeviceRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasDevice(c) {
			return fiber.ErrUnauthorized
		}

		return c.Next()
	}
}

// WithDevice is a decorator that provides the current device to the handler.
// It assumes that the device is stored in the Locals under the key LocalsDevice.
// If the device is not present, it will panic.
//
// It is a convenience function that wraps the call to GetDevice and calls the
// handler with the device as the first argument.
func WithDevice(handler func(models.Device, *fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return handler(GetDevice(c), c)
	}
}
