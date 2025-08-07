package e2e

import (
	"encoding/json"
	"testing"

	"github.com/go-resty/resty/v2"
)

type messageState struct {
	ID          string   `json:"id"`
	DeviceID    string   `json:"deviceId"`
	State       string   `json:"state"`
	IsHashed    bool     `json:"isHashed"`
	IsEncrypted bool     `json:"isEncrypted"`
	Recipients  []string `json:"recipients"`
	States      []state  `json:"states"`
}

type state struct {
	PhoneNumber string `json:"phoneNumber"`
	State       string `json:"state"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func TestMessages_GetMessages(t *testing.T) {
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
			name: "Successful retrieval with default parameters",
			setup: func() {
				// Test data is populated by the test infrastructure
			},
			expectedStatusCode: 200,
			request: func() *resty.Request {
				return authorizedClient.R()
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 200 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var result []messageState
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				// Verify response structure
				if len(result) > 0 {
					msg := result[0]
					if msg.ID == "" {
						t.Error("message ID is empty")
					}
					if msg.DeviceID == "" {
						t.Error("device ID is empty")
					}
					if msg.State == "" {
						t.Error("message state is empty")
					}
				}

				// Verify response headers
				if response.Header().Get("Content-Type") != "application/json" {
					t.Error("expected Content-Type to be application/json")
				}
			},
		},
		{
			name: "Pagination with limit=10, offset=5",
			setup: func() {
				// Test data is populated by the test infrastructure
			},
			expectedStatusCode: 200,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetQueryParams(map[string]string{
						"limit":  "10",
						"offset": "5",
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 200 {
					t.Fatal(response.StatusCode(), response.String())
				}

				// Verify X-Total-Count header
				totalCount := response.Header().Get("X-Total-Count")
				if totalCount == "" {
					t.Error("expected X-Total-Count header")
				}

				var result []messageState
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				// Verify pagination limits
				if len(result) > 10 {
					t.Errorf("expected at most 10 messages, got %d", len(result))
				}
			},
		},
		{
			name: "Date range filter",
			setup: func() {
				// Test data is populated by the test infrastructure
			},
			expectedStatusCode: 200,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetQueryParams(map[string]string{
						"from": "2025-07-01T00:00:00Z",
						"to":   "2025-07-31T23:59:59Z",
					})
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 200 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var result []messageState
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				// Verify response structure
				for _, msg := range result {
					if msg.ID == "" {
						t.Error("message ID is empty")
					}
					if msg.DeviceID == "" {
						t.Error("device ID is empty")
					}
				}
			},
		},
		{
			name: "State filter (Sent)",
			setup: func() {
				// Test data is populated by the test infrastructure
			},
			expectedStatusCode: 200,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetQueryParam("state", "Sent")
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 200 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var result []messageState
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				// Verify all messages have Sent state
				for _, msg := range result {
					if msg.State != "Sent" {
						t.Errorf("expected state 'Sent', got '%s'", msg.State)
					}
				}
			},
		},
		{
			name: "Device ID filter",
			setup: func() {
				// Test data is populated by the test infrastructure
			},
			expectedStatusCode: 200,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetQueryParam("deviceId", credentials.ID)
			},
			validate: func(t *testing.T, response *resty.Response) {
				if response.StatusCode() != 200 {
					t.Fatal(response.StatusCode(), response.String())
				}

				var result []messageState
				if err := json.Unmarshal(response.Body(), &result); err != nil {
					t.Fatal(err)
				}

				// Verify response structure
				for _, msg := range result {
					if msg.ID == "" {
						t.Error("message ID is empty")
					}
					if msg.DeviceID == "" {
						t.Error("device ID is empty")
					}
				}
			},
		},
		{
			name: "Invalid date format",
			setup: func() {
				// Test data is populated by the test infrastructure
			},
			expectedStatusCode: 400,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetQueryParam("from", "invalid")
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
			name: "Invalid state value",
			setup: func() {
				// Test data is populated by the test infrastructure
			},
			expectedStatusCode: 400,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetQueryParam("state", "InvalidState")
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
			name: "Invalid device ID length",
			setup: func() {
				// Test data is populated by the test infrastructure
			},
			expectedStatusCode: 400,
			request: func() *resty.Request {
				return authorizedClient.R().
					SetQueryParam("deviceId", "invalid_length_device_id")
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
				// Test data is populated by the test infrastructure
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
				// Test data is populated by the test infrastructure
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

			res, err := c.request().Get("messages")
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
