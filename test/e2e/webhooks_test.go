package e2e

import (
	"encoding/json"
	"testing"

	"github.com/go-resty/resty/v2"
)

type webhook struct {
	ID       string `json:"id"`
	DeviceID string `json:"deviceId,omitempty"`
	URL      string `json:"url"`
	Event    string `json:"event"`
}

func TestWebhooks_Get(t *testing.T) {
	credentials := mobileDeviceRegister(t, publicMobileClient)
	authorizedClient := publicUserClient.Clone().SetBasicAuth(credentials.Login, credentials.Password)

	cases := []struct {
		name               string
		setup              func()
		expectedStatusCode int
		request            func() *resty.Request
		validate           func(t *testing.T, response *resty.Response)
	}{
		{
			name: "Successful retrieval of empty webhook list",
			setup: func() {
				// Start with empty webhook list
			},
			expectedStatusCode: 200,
			request: func() *resty.Request {
				return authorizedClient.R()
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 200 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var result []webhook
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				// Verify response structure
				if len(result) != 0 {
					t.Errorf("expected empty webhook list, got %d webhooks", len(result))
				}

				// Verify response headers
				if response.Header().Get("Content-Type") != "application/json" {
					t.Error("expected Content-Type to be application/json")
				}
			},
		},
		{
			name: "List webhooks after creation",
			setup: func() {
				// Create a webhook first
				_, err := authorizedClient.R().
					SetBody(webhook{
						URL:   "https://example.com/list-test",
						Event: "sms:delivered",
					}).Post("webhooks")
				if err != nil {
					t.Fatal(err)
				}
			},
			expectedStatusCode: 200,
			request: func() *resty.Request {
				return authorizedClient.R()
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 200 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var result []webhook
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				// Verify response structure
				if len(result) == 0 {
					t.Error("expected webhook list to contain created webhooks")
				}

				// Verify webhook structure
				for _, webhook := range result {
					if webhook.ID == "" {
						t.Error("webhook ID is empty")
					}
					if webhook.URL == "" {
						t.Error("webhook URL is empty")
					}
					if webhook.Event == "" {
						t.Error("webhook event is empty")
					}
				}
			},
		},
		{
			name: "Missing authentication",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 401,
			request: func() *resty.Request {
				return publicUserClient.R()
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 401 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var err errorResponse
				if err := json.Unmarshal(response.Body(), &err); err != nil {
					t.Fatal(err)
				}

				if err.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
		{
			name: "Invalid credentials",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 401,
			request: func() *resty.Request {
				return publicUserClient.R().SetBasicAuth("invalid", "credentials")
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 401 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var err errorResponse
				if err := json.Unmarshal(response.Body(), &err); err != nil {
					t.Fatal(err)
				}

				if err.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.setup()

			res, err := c.request().Get("webhooks")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}

			if c.validate != nil {
				c.validate(t, res)
			}
		})
	}
}

func TestWebhooks_Post(t *testing.T) {
	credentials := mobileDeviceRegister(t, publicMobileClient)
	authorizedClient := publicUserClient.Clone().SetBasicAuth(credentials.Login, credentials.Password)

	cases := []struct {
		name               string
		setup              func()
		expectedStatusCode int
		request            func() *resty.Request
		validate           func(t *testing.T, response *resty.Response)
	}{
		{
			name: "Create webhook with valid data",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 201,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetBody(webhook{
						URL:   "https://example.com/webhook",
						Event: "sms:received",
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 201 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var result webhook
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() {
					_, err := authorizedClient.R().Delete("webhooks/" + result.ID)
					if err != nil {
						t.Error(err)
					}
				})

				// Verify response structure
				if result.ID == "" {
					t.Error("webhook ID is empty")
				}
				if result.URL != "https://example.com/webhook" {
					t.Errorf("expected URL 'https://example.com/webhook', got '%s'", result.URL)
				}
				if result.Event != "sms:received" {
					t.Errorf("expected event 'sms:received', got '%s'", result.Event)
				}
			},
		},
		{
			name: "Create webhook with device_id",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 201,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetBody(webhook{
						DeviceID: credentials.ID,
						URL:      "https://example.com/device-webhook",
						Event:    "sms:sent",
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 201 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var result webhook
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() {
					_, err := authorizedClient.R().Delete("webhooks/" + result.ID)
					if err != nil {
						t.Error(err)
					}
				})

				// Verify response structure
				if result.ID == "" {
					t.Error("webhook ID is empty")
				}
				if result.DeviceID != credentials.ID {
					t.Errorf("expected device_id '%s', got '%s'", credentials.ID, result.DeviceID)
				}
				if result.URL != "https://example.com/device-webhook" {
					t.Errorf("expected URL 'https://example.com/device-webhook', got '%s'", result.URL)
				}
				if result.Event != "sms:sent" {
					t.Errorf("expected event 'sms:sent', got '%s'", result.Event)
				}
			},
		},
		{
			name: "Create webhook with different event types",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 201,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetBody(webhook{
						URL:   "https://example.com/data-webhook",
						Event: "sms:data-received",
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 201 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var result webhook
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() {
					_, err := authorizedClient.R().Delete("webhooks/" + result.ID)
					if err != nil {
						t.Error(err)
					}
				})

				// Verify response structure
				if result.Event != "sms:data-received" {
					t.Errorf("expected event 'sms:data-received', got '%s'", result.Event)
				}
			},
		},
		{
			name: "Invalid URL format",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 400,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetBody(webhook{
						URL:   "invalid-url",
						Event: "sms:received",
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 400 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var err errorResponse
				if err := json.Unmarshal(response.Body(), &err); err != nil {
					t.Fatal(err)
				}

				if err.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
		{
			name: "Missing required fields",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 400,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetBody(webhook{
						// Missing URL and Event
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 400 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var err errorResponse
				if err := json.Unmarshal(response.Body(), &err); err != nil {
					t.Fatal(err)
				}

				if err.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
		{
			name: "Invalid event type",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 400,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetBody(webhook{
						URL:   "https://example.com/webhook",
						Event: "invalid:event",
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 400 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var err errorResponse
				if err := json.Unmarshal(response.Body(), &err); err != nil {
					t.Fatal(err)
				}

				if err.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
		{
			name: "Invalid device_id length",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 400,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetBody(webhook{
						DeviceID: "invalid_length_device_id",
						URL:      "https://example.com/webhook",
						Event:    "sms:received",
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 400 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var err errorResponse
				if err := json.Unmarshal(response.Body(), &err); err != nil {
					t.Fatal(err)
				}

				if err.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
		{
			name: "Missing authentication",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 401,
			request: func() *resty.Request {
				return publicUserClient.R().
					SetBody(webhook{
						URL:   "https://example.com/webhook",
						Event: "sms:received",
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 401 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var err errorResponse
				if err := json.Unmarshal(response.Body(), &err); err != nil {
					t.Fatal(err)
				}

				if err.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
		{
			name: "Invalid credentials",
			setup: func() {
				// No setup needed
			},
			expectedStatusCode: 401,
			request: func() *resty.Request {
				return publicUserClient.R().
					SetBody(webhook{
						URL:   "https://example.com/webhook",
						Event: "sms:received",
					}).SetBasicAuth("invalid", "credentials")
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 401 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var err errorResponse
				if err := json.Unmarshal(response.Body(), &err); err != nil {
					t.Fatal(err)
				}

				if err.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.setup()

			res, err := c.request().Post("webhooks")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}

			if c.validate != nil {
				c.validate(t, res)
			}
		})
	}
}

func TestWebhooks_Delete(t *testing.T) {
	credentials := mobileDeviceRegister(t, publicMobileClient)
	authorizedClient := publicUserClient.Clone().SetBasicAuth(credentials.Login, credentials.Password)

	cases := []struct {
		name               string
		setup              func() string
		expectedStatusCode int
		request            func(id string) *resty.Request
		validate           func(t *testing.T, response *resty.Response)
	}{
		{
			name: "Remove webhook by ID",
			setup: func() string {
				// Create a webhook to delete (server will generate ID)
				resp, err := authorizedClient.R().
					SetBody(webhook{
						URL:   "https://example.com/delete-test",
						Event: "sms:failed",
					}).Post("webhooks")
				if err != nil {
					t.Fatal(err)
				}

				var created webhook
				if err := json.Unmarshal(resp.Body(), &created); err != nil {
					t.Fatal(err)
				}

				return created.ID
			},
			expectedStatusCode: 204,
			request: func(id string) *resty.Request {
				return authorizedClient.R().SetPathParam("id", id)
			},
			validate: func(t *testing.T, response *resty.Response) {
				if len(response.Body()) != 0 {
					t.Error("expected empty response body for 204 status")
				}
			},
		},
		{
			name: "Remove non-existent webhook",
			setup: func() string {
				// No setup needed
				return ""
			},
			expectedStatusCode: 204,
			request: func(id string) *resty.Request {
				return authorizedClient.R().SetPathParam("id", "non-existent-id")
			},
			validate: func(t *testing.T, response *resty.Response) {
				if len(response.Body()) != 0 {
					t.Error("expected empty response body for 204 status")
				}
			},
		},
		{
			name: "Missing authentication",
			setup: func() string {
				// No setup needed
				return ""
			},
			expectedStatusCode: 401,
			request: func(id string) *resty.Request {
				return publicUserClient.R().SetPathParam("id", "test-id")
			},
			validate: func(t *testing.T, response *resty.Response) {
				var errResp errorResponse
				if err := json.Unmarshal(response.Body(), &errResp); err != nil {
					t.Fatal(err)
				}

				if errResp.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
		{
			name: "Invalid credentials",
			setup: func() string {
				// No setup needed
				return ""
			},
			expectedStatusCode: 401,
			request: func(id string) *resty.Request {
				return publicUserClient.R().SetBasicAuth("invalid", "credentials").SetPathParam("id", "test-id")
			},
			validate: func(t *testing.T, response *resty.Response) {
				var errResp errorResponse
				if err := json.Unmarshal(response.Body(), &errResp); err != nil {
					t.Fatal(err)
				}

				if errResp.Message == "" {
					t.Error("expected error message in response")
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			webhookID := c.setup()
			// Clean up the webhook if one was created
			if webhookID != "" {
				t.Cleanup(func() {
					authorizedClient.R().Delete("webhooks/" + webhookID)
				})
			}

			res, err := c.request(webhookID).Delete("webhooks/{id}")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}

			if c.validate != nil {
				c.validate(t, res)
			}

		})
	}
}
