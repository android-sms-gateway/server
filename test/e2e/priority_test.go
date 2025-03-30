package e2e

import (
	"testing"

	"github.com/capcom6/go-helpers/anys"
)

func TestPriorityPost(t *testing.T) {
	cases := []struct {
		name               string
		priority           *int
		expectedStatusCode int
	}{
		{
			name:               "min priority",
			priority:           anys.AsPointer(-128),
			expectedStatusCode: 202,
		},
		{
			name:               "max priority",
			priority:           anys.AsPointer(127),
			expectedStatusCode: 202,
		},
		{
			name:               "invalid priority",
			priority:           anys.AsPointer(128),
			expectedStatusCode: 400,
		},
		{
			name:               "invalid priority",
			priority:           anys.AsPointer(-129),
			expectedStatusCode: 400,
		},
		{
			name:               "default priority",
			priority:           nil,
			expectedStatusCode: 202,
		},
	}

	req := map[string]any{
		"message": "test",
		"phoneNumbers": []string{
			"+79999999999",
		},
	}

	credentials := mobileDeviceRegister(t, publicMobileClient)
	client := publicUserClient.SetBasicAuth(credentials.Login, credentials.Password)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.priority != nil {
				req["priority"] = *c.priority
			} else {
				delete(req, "priority")
			}
			res, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(req).
				Post("messages")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}
		})
	}
}
