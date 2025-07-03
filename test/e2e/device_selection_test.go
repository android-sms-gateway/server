package e2e

import (
	"testing"

	"github.com/capcom6/go-helpers/anys"
)

func TestDeviceSelection(t *testing.T) {
	// Register first device
	firstDevice := mobileDeviceRegister(t, publicMobileClient)
	client := publicUserClient.SetBasicAuth(firstDevice.Login, firstDevice.Password)

	// Register a second device to test explicit device selection
	secondDevice := mobileDeviceRegister(
		t,
		publicMobileClient,
		(&mobileDeviceRegisterOptions{}).
			withCredentials(firstDevice.Login, firstDevice.Password),
	)

	cases := []struct {
		name               string
		deviceID           *string
		expectedStatusCode int
	}{
		{
			name:               "explicit device selection",
			deviceID:           anys.AsPointer(secondDevice.ID),
			expectedStatusCode: 202,
		},
		{
			name:               "invalid device ID",
			deviceID:           anys.AsPointer("invalid-device-id"),
			expectedStatusCode: 400,
		},
		{
			name:               "no device ID (random selection)",
			deviceID:           nil,
			expectedStatusCode: 202,
		},
	}

	req := map[string]any{
		"message": "test",
		"phoneNumbers": []string{
			"+79999999999",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.deviceID != nil {
				req["deviceId"] = *c.deviceID
			} else {
				delete(req, "deviceId")
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
