package e2e

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

type messageState struct {
	ID          string               `json:"id"`
	DeviceID    string               `json:"deviceId"`
	State       string               `json:"state"`
	IsHashed    bool                 `json:"isHashed"`
	IsEncrypted bool                 `json:"isEncrypted"`
	Recipients  []recipientState     `json:"recipients"`
	States      map[string]time.Time `json:"states"`
	CreatedAt   string               `json:"createdAt"`
}

type recipientState struct {
	PhoneNumber string `json:"phoneNumber"`
	State       string `json:"state"`
	Error       string `json:"error"`
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

func TestMessages_Post(t *testing.T) {
	credentials := mobileDeviceRegister(t, publicMobileClient)
	authorizedClient := publicUserClient.Clone().SetBasicAuth(credentials.Login, credentials.Password)

	cases := []struct {
		name               string
		req                map[string]any
		expectedStatusCode int
		expectedMessage    string
	}{
		{
			name: "duplicate phone numbers are rejected",
			req: map[string]any{
				"message": "test",
				"phoneNumbers": []string{
					"+79999999999",
					"+79999999999",
				},
			},
			expectedStatusCode: 400,
			expectedMessage:    "failed to validate: Key: 'Message.PhoneNumbers' Error:Field validation for 'PhoneNumbers' failed on the 'unique' tag",
		},
		{
			name: "distinct phone numbers are accepted",
			req: map[string]any{
				"message": "test",
				"phoneNumbers": []string{
					"+79999999999",
					"+79990001234",
				},
			},
			expectedStatusCode: 202,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := authorizedClient.R().
				SetHeader("Content-Type", "application/json").
				SetBody(c.req).
				Post("messages")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}

			if c.expectedMessage != "" {
				var resp errorResponse
				if err := json.Unmarshal(res.Body(), &resp); err != nil {
					t.Fatal(err)
				}

				if resp.Message != c.expectedMessage {
					t.Errorf("expected message %q, got %q", c.expectedMessage, resp.Message)
				}
			}
		})
	}
}

func TestMessages_PostDuplicateMessageID(t *testing.T) {
	credentials := mobileDeviceRegister(t, publicMobileClient)
	authorizedClient := publicUserClient.Clone().SetBasicAuth(credentials.Login, credentials.Password)

	// The message ID uniqueness scope is per device (unq_messages_id_device),
	// so pin the device to guarantee the ExtID collision.
	base := map[string]any{
		"id":           "e2e-duplicate-message-id",
		"message":      "test",
		"deviceId":     credentials.ID,
		"phoneNumbers": []string{"+79999999999"},
	}

	t.Run("first message with the ID is accepted", func(t *testing.T) {
		res, err := authorizedClient.R().
			SetHeader("Content-Type", "application/json").
			SetBody(base).
			Post("messages")
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode() != 202 {
			t.Fatal(res.StatusCode(), res.String())
		}
	})

	t.Run("duplicate message ID is rejected", func(t *testing.T) {
		req := map[string]any{
			"id":           base["id"],
			"message":      "test",
			"deviceId":     credentials.ID,
			"phoneNumbers": []string{"+79990001234"},
		}

		res, err := authorizedClient.R().
			SetHeader("Content-Type", "application/json").
			SetBody(req).
			Post("messages")
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode() != 409 {
			t.Fatal(res.StatusCode(), res.String())
		}

		var resp errorResponse
		if err := json.Unmarshal(res.Body(), &resp); err != nil {
			t.Fatal(err)
		}

		if resp.Message != "message with the same ID already exists" {
			t.Errorf("expected message %q, got %q", "Message with the same ID already exists", resp.Message)
		}
	})
}

// TestMessages_CreatedAt pins the creation time reported on enqueue to the value stored in the
// database. The enqueued state is built from the freshly inserted model, so it only agrees with a
// later read when `created_at` is read back from the row. GORM reads such DB-generated columns
// through RETURNING, which its MySQL driver enables for MariaDB only - without an explicit read the
// accepted response reports the zero time on MySQL and the cached state carries it until it expires.
func TestMessages_CreatedAt(t *testing.T) {
	credentials := mobileDeviceRegister(t, publicMobileClient)
	authorizedClient := publicUserClient.Clone().SetBasicAuth(credentials.Login, credentials.Password)

	res, err := authorizedClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]any{
			"id":           "e2e-created-at",
			"message":      "test",
			"deviceId":     credentials.ID,
			"phoneNumbers": []string{"+79990001234"},
		}).
		Post("messages")
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode() != 202 {
		t.Fatal(res.StatusCode(), res.String())
	}

	var enqueued messageState
	if err := json.Unmarshal(res.Body(), &enqueued); err != nil {
		t.Fatal(err)
	}

	createdAt, err := time.Parse(time.RFC3339, enqueued.CreatedAt)
	if err != nil {
		t.Fatalf("createdAt %q is not an RFC 3339 timestamp: %v", enqueued.CreatedAt, err)
	}

	if createdAt.IsZero() {
		t.Error("createdAt reports the zero time")
	}

	// The enqueue path caches the state it just built, so compare against the history endpoint,
	// which reads the row from the database instead of the cache.
	getRes, err := authorizedClient.R().SetQueryParam("deviceId", credentials.ID).Get("messages")
	if err != nil {
		t.Fatal(err)
	}

	if getRes.StatusCode() != 200 {
		t.Fatal(getRes.StatusCode(), getRes.String())
	}

	var listed []messageState
	if err := json.Unmarshal(getRes.Body(), &listed); err != nil {
		t.Fatal(err)
	}

	var stored *messageState
	for i, msg := range listed {
		if msg.ID == enqueued.ID {
			stored = &listed[i]

			break
		}
	}

	if stored == nil {
		t.Fatalf("message %q is missing from the history response", enqueued.ID)
	}

	if stored.CreatedAt != enqueued.CreatedAt {
		t.Errorf("createdAt stored as %q does not match enqueued %q", stored.CreatedAt, enqueued.CreatedAt)
	}
}
