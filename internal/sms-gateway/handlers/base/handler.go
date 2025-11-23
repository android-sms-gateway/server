package base

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Validatable interface {
	Validate() error
}

type Handler struct {
	Logger    *zap.Logger
	Validator *validator.Validate
}

func (h *Handler) BodyParserValidator(c *fiber.Ctx, out any) error {
	if err := c.BodyParser(out); err != nil {
		return fmt.Errorf("failed to parse body: %w", err)
	}

	return h.ValidateStruct(out)
}

func (h *Handler) QueryParserValidator(c *fiber.Ctx, out any) error {
	if err := c.QueryParser(out); err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	return h.ValidateStruct(out)
}

func (h *Handler) ParamsParserValidator(c *fiber.Ctx, out any) error {
	if err := c.ParamsParser(out); err != nil {
		return fmt.Errorf("failed to parse params: %w", err)
	}

	return h.ValidateStruct(out)
}

func (h *Handler) ValidateStruct(out any) error {
	if h.Validator != nil {
		if err := h.Validator.Var(out, "required,dive"); err != nil {
			return fmt.Errorf("failed to validate: %w", err)
		}
	}

	if req, ok := out.(Validatable); ok {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("failed to validate: %w", err)
		}
	}

	return nil
}
