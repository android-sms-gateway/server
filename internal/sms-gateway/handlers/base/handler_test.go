package base_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap/zaptest"
)

type testRequestBody struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age"  validate:"required"`
}

type testRequestBodyNoValidate struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age"  validate:"required"`
}

type testRequestQuery struct {
	Name string `query:"name" validate:"required"`
	Age  int    `query:"age"  validate:"required"`
}

type testRequestParams struct {
	ID   string `params:"id"   validate:"required"`
	Name string `params:"name" validate:"required"`
}

func (t *testRequestBody) Validate() error {
	if t.Age < 18 {
		return errors.New("must be at least 18 years old")
	}
	return nil
}

func (t *testRequestQuery) Validate() error {
	if t.Age < 18 {
		return errors.New("must be at least 18 years old")
	}
	return nil
}

func (t *testRequestParams) Validate() error {
	if t.ID == "invalid" {
		return errors.New("invalid ID")
	}
	return nil
}

func TestHandler_BodyParserValidator(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validate := validator.New()

	handler := &base.Handler{
		Logger:    logger,
		Validator: validate,
	}

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		var body testRequestBody
		return handler.BodyParserValidator(c, &body)
	})
	app.Post("/test2", func(c *fiber.Ctx) error {
		var body testRequestBodyNoValidate
		return handler.BodyParserValidator(c, &body)
	})

	tests := []struct {
		description    string
		path           string
		payload        any
		expectedStatus int
	}{
		{
			description:    "Valid request body",
			path:           "/test",
			payload:        &testRequestBody{Name: "John Doe", Age: 25},
			expectedStatus: fiber.StatusOK,
		},
		{
			description:    "Invalid request body - missing name",
			path:           "/test",
			payload:        &testRequestBody{Age: 25},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			description:    "Invalid request body - age too low",
			path:           "/test",
			payload:        &testRequestBody{Name: "John Doe", Age: 17},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			description:    "Valid request body - no validation",
			path:           "/test2",
			payload:        &testRequestBodyNoValidate{Name: "John Doe", Age: 17},
			expectedStatus: fiber.StatusOK,
		},
		{
			description:    "No request body",
			path:           "/test",
			payload:        nil,
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var req *http.Request
			if test.payload != nil {
				bodyBytes, _ := json.Marshal(test.payload)
				req = httptest.NewRequest(http.MethodPost, test.path, bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(http.MethodPost, test.path, nil)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test failed: %v", err)
			}
			if test.expectedStatus != resp.StatusCode {
				t.Errorf("Expected status code %d, got %d", test.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestHandler_QueryParserValidator(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validate := validator.New()

	handler := &base.Handler{
		Logger:    logger,
		Validator: validate,
	}

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		var query testRequestQuery
		return handler.QueryParserValidator(c, &query)
	})

	tests := []struct {
		description    string
		path           string
		expectedStatus int
	}{
		{
			description:    "Invalid query parameters - non-integer age",
			path:           "/test?name=John&age=abc",
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			description:    "Valid query parameters",
			path:           "/test?name=John&age=25",
			expectedStatus: fiber.StatusOK,
		},
		{
			description:    "Invalid query parameters - missing name",
			path:           "/test?age=25",
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			description:    "Invalid query parameters - age too low",
			path:           "/test?name=John&age=17",
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			description:    "Invalid query parameters - missing age",
			path:           "/test?name=John",
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.path, nil)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test failed: %v", err)
			}
			if test.expectedStatus != resp.StatusCode {
				t.Errorf("Expected status code %d, got %d", test.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestHandler_ParamsParserValidator(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validate := validator.New()

	handler := &base.Handler{
		Logger:    logger,
		Validator: validate,
	}

	app := fiber.New()
	app.Get("/test/:id/:name", func(c *fiber.Ctx) error {
		var params testRequestParams
		return handler.ParamsParserValidator(c, &params)
	})

	tests := []struct {
		description    string
		path           string
		expectedStatus int
	}{
		{
			description:    "Valid path parameters",
			path:           "/test/123/John",
			expectedStatus: fiber.StatusOK,
		},
		{
			description:    "Invalid path parameters - missing id",
			path:           "/test//John",
			expectedStatus: fiber.StatusNotFound,
		},
		{
			description:    "Invalid path parameters - missing name",
			path:           "/test/123/",
			expectedStatus: fiber.StatusNotFound,
		},
		{
			description:    "Invalid path parameters - invalid ID",
			path:           "/test/invalid/John",
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.path, nil)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test failed: %v", err)
			}
			if test.expectedStatus != resp.StatusCode {
				t.Errorf("Expected status code %d, got %d", test.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestHandler_ValidateStruct(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validate := validator.New()

	// Test with validator
	handlerWithValidator := &base.Handler{
		Logger:    logger,
		Validator: validate,
	}

	// Test without validator
	handlerWithoutValidator := &base.Handler{
		Logger:    logger,
		Validator: nil,
	}

	tests := []struct {
		description    string
		handler        *base.Handler
		input          any
		expectedStatus int
	}{
		{
			description:    "Valid struct with validator",
			handler:        handlerWithValidator,
			input:          &testRequestBody{Name: "John Doe", Age: 25},
			expectedStatus: fiber.StatusOK,
		},
		{
			description:    "Invalid struct with validator - missing required field",
			handler:        handlerWithValidator,
			input:          &testRequestBody{Age: 25},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			description:    "Invalid struct with validator - custom validation fails",
			handler:        handlerWithValidator,
			input:          &testRequestBody{Name: "John Doe", Age: 17},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			description:    "Valid struct without validator",
			handler:        handlerWithoutValidator,
			input:          &testRequestBody{Name: "John Doe", Age: 25},
			expectedStatus: fiber.StatusOK,
		},
		{
			description:    "Invalid struct without validator - custom validation fails",
			handler:        handlerWithoutValidator,
			input:          &testRequestBody{Name: "John Doe", Age: 17},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			description:    "Valid struct with Validatable interface",
			handler:        handlerWithValidator,
			input:          &testRequestQuery{Name: "John", Age: 25},
			expectedStatus: fiber.StatusOK,
		},
		{
			description:    "Invalid struct with Validatable interface",
			handler:        handlerWithValidator,
			input:          &testRequestQuery{Name: "John", Age: 17},
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.handler.ValidateStruct(test.input)

			if test.expectedStatus == fiber.StatusOK && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if test.expectedStatus == fiber.StatusBadRequest && err == nil {
				t.Errorf("Expected error, got nil")
			}
		})
	}
}
