//nolint:testpackage // errorHandler is unexported; in-package test required.
package messages

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	messages "github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestThirdPartyController() *ThirdPartyController {
	return &ThirdPartyController{
		Handler: base.Handler{
			Logger: zap.NewNop(),
		},
	}
}

func TestThirdPartyControllerErrorHandler(t *testing.T) {
	const enqueueWrap = "failed to enqueue message: %w"

	tests := []struct {
		name       string
		handlerErr error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "wrapped validation error emits its own text (no wrap prefix)",
			handlerErr: fmt.Errorf(enqueueWrap, messages.ValidationError("phone numbers must be unique")),
			wantStatus: fiber.StatusBadRequest,
			wantBody:   `{"message":"phone numbers must be unique"}`,
		},
		{
			name:       "wrapped duplicate ID emits sentinel text (no wrap prefix)",
			handlerErr: fmt.Errorf(enqueueWrap, messages.ErrMessageAlreadyExists),
			wantStatus: fiber.StatusConflict,
			wantBody:   `{"message":"message with the same ID already exists"}`,
		},
		{
			name:       "wrapped multiple messages found keeps err.Error() text (unchanged)",
			handlerErr: fmt.Errorf(enqueueWrap, messages.ErrMultipleMessagesFound),
			wantStatus: fiber.StatusBadRequest,
			wantBody:   `{"message":"failed to enqueue message: multiple messages found"}`,
		},
		{
			name:       "wrapped no content keeps err.Error() text (unchanged)",
			handlerErr: fmt.Errorf(enqueueWrap, messages.ErrNoContent),
			wantStatus: fiber.StatusBadRequest,
			wantBody:   `{"message":"failed to enqueue message: no text or data content"}`,
		},
		{
			name:       "bare validation error keeps its own text",
			handlerErr: messages.ValidationError("invalid phone number"),
			wantStatus: fiber.StatusBadRequest,
			wantBody:   `{"message":"invalid phone number"}`,
		},
		{
			name:       "wrapped validation error variation",
			handlerErr: fmt.Errorf(enqueueWrap, messages.ValidationError("not mobile phone number")),
			wantStatus: fiber.StatusBadRequest,
			wantBody:   `{"message":"not mobile phone number"}`,
		},
		{
			name:       "zero-value validation error emits its own empty text (no wrap prefix)",
			handlerErr: fmt.Errorf(enqueueWrap, messages.ValidationError("")),
			wantStatus: fiber.StatusBadRequest,
			wantBody:   `{"message":""}`,
		},
		{
			name:       "not found keeps 404",
			handlerErr: fmt.Errorf(enqueueWrap, messages.ErrMessageNotFound),
			wantStatus: fiber.StatusNotFound,
			wantBody:   `{"message":"failed to enqueue message: message not found"}`,
		},
		{
			name:       "queue limit keeps 503",
			handlerErr: fmt.Errorf(enqueueWrap, messages.ErrQueueLimitExceeded),
			wantStatus: fiber.StatusServiceUnavailable,
			wantBody:   `{"message":"failed to enqueue message: queue limits exceeded"}`,
		},
		{
			name:       "not pending keeps 409",
			handlerErr: fmt.Errorf(enqueueWrap, messages.ErrMessageNotPending),
			wantStatus: fiber.StatusConflict,
			wantBody:   `{"message":"failed to enqueue message: message is not pending"}`,
		},
		{
			name:       "device not found keeps 400",
			handlerErr: fmt.Errorf("failed to select device: %w", devices.ErrNotFound),
			wantStatus: fiber.StatusBadRequest,
			wantBody:   `{"message":"failed to select device: record not found"}`,
		},
		{
			name:       "unknown error keeps 500",
			handlerErr: errors.New("boom"),
			wantStatus: fiber.StatusInternalServerError,
			wantBody:   `{"message":"failed to handle request"}`,
		},
		{
			name:       "fiber error passes through unchanged",
			handlerErr: fmt.Errorf("wrapped: %w", fiber.NewError(fiber.StatusTeapot, "teapot")),
			wantStatus: fiber.StatusTeapot,
			wantBody:   `{"message":"teapot"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestThirdPartyController()

			app := fiber.New(fiber.Config{
				// Mirrors go-infra-fx/http errorHandler ({"message": err.Error()}).
				ErrorHandler: func(c *fiber.Ctx, err error) error {
					code := fiber.StatusInternalServerError
					var fiberErr *fiber.Error
					if errors.As(err, &fiberErr) {
						code = fiberErr.Code
					}
					return c.Status(code).JSON(&fiber.Map{"message": err.Error()})
				},
			})
			app.Use(h.errorHandler)
			app.Post("/enqueue", func(_ *fiber.Ctx) error {
				return tt.handlerErr
			})

			req := httptest.NewRequest(http.MethodPost, "/enqueue", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tt.wantStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, tt.wantBody, string(body))
		})
	}
}
