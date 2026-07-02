package e2e

import (
	"testing"
)

func TestInboxRefresh_WebhookDelivery(t *testing.T) {
	credentials := mobileDeviceRegister(t, publicMobileClient)
	authorizedClient := publicUserClient.Clone().SetBasicAuth(credentials.Login, credentials.Password)

	type inboxRefreshRequest struct {
		Since           string  `json:"since"`
		Until           string  `json:"until"`
		WebhookDelivery *string `json:"webhookDelivery,omitempty"`
		TriggerWebhooks *bool   `json:"triggerWebhooks,omitempty"`
	}

	cases := []struct {
		name               string
		body               inboxRefreshRequest
		expectedStatusCode int
	}{
		{
			name: "webhookDelivery Individual",
			body: inboxRefreshRequest{
				Since:           "2024-01-01T00:00:00Z",
				Until:           "2024-01-01T23:59:59Z",
				WebhookDelivery: ptr("Individual"),
			},
			expectedStatusCode: 202,
		},
		{
			name: "webhookDelivery Batch",
			body: inboxRefreshRequest{
				Since:           "2024-01-01T00:00:00Z",
				Until:           "2024-01-01T23:59:59Z",
				WebhookDelivery: ptr("Batch"),
			},
			expectedStatusCode: 202,
		},
		{
			name: "webhookDelivery Disabled",
			body: inboxRefreshRequest{
				Since:           "2024-01-01T00:00:00Z",
				Until:           "2024-01-01T23:59:59Z",
				WebhookDelivery: ptr("Disabled"),
			},
			expectedStatusCode: 202,
		},
		{
			name: "legacy triggerWebhooks true",
			body: inboxRefreshRequest{
				Since:           "2024-01-01T00:00:00Z",
				Until:           "2024-01-01T23:59:59Z",
				TriggerWebhooks: ptr(true),
			},
			expectedStatusCode: 202,
		},
		{
			name: "legacy triggerWebhooks false",
			body: inboxRefreshRequest{
				Since:           "2024-01-01T00:00:00Z",
				Until:           "2024-01-01T23:59:59Z",
				TriggerWebhooks: ptr(false),
			},
			expectedStatusCode: 202,
		},
		{
			name: "webhookDelivery takes precedence over triggerWebhooks",
			body: inboxRefreshRequest{
				Since:           "2024-01-01T00:00:00Z",
				Until:           "2024-01-01T23:59:59Z",
				WebhookDelivery: ptr("Batch"),
				TriggerWebhooks: ptr(false),
			},
			expectedStatusCode: 202,
		},
		{
			name: "neither webhookDelivery nor triggerWebhooks provided",
			body: inboxRefreshRequest{
				Since: "2024-01-01T00:00:00Z",
				Until: "2024-01-01T23:59:59Z",
			},
			expectedStatusCode: 202,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := authorizedClient.R().
				SetBody(c.body).
				Post("inbox/refresh")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatalf("expected %d, got %d: %s", c.expectedStatusCode, res.StatusCode(), res.String())
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
