package e2e

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

type mobileRegisterResponse struct {
	Token    string `json:"token"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func TestPublicDeviceRegister(t *testing.T) {
	cases := []struct {
		name               string
		headers            map[string]string
		expectedStatusCode int
	}{
		{
			name: "with valid token",
			headers: map[string]string{
				"Authorization": "Bearer 123456789",
			},
			expectedStatusCode: 201,
		},
		{
			name:               "without token",
			headers:            map[string]string{},
			expectedStatusCode: 201,
		},
		{
			name: "with invalid token",
			headers: map[string]string{
				"Authorization": "Bearer 987654321",
			},
			expectedStatusCode: 201,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := publicMobileClient.R().
				SetHeader("Content-Type", "application/json").
				SetBody(`{"name": "Public Device Name", "pushToken": "token"}`).
				SetHeaders(c.headers).
				Post("device")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}
		})
	}
}

func TestPrivateDeviceRegister(t *testing.T) {
	cases := []struct {
		name               string
		headers            map[string]string
		expectedStatusCode int
	}{
		{
			name: "with valid token",
			headers: map[string]string{
				"Authorization": "Bearer 123456789",
			},
			expectedStatusCode: 201,
		},
		{
			name:               "without token",
			headers:            map[string]string{},
			expectedStatusCode: 401,
		},
		{
			name: "with invalid token",
			headers: map[string]string{
				"Authorization": "Bearer 987654321",
			},
			expectedStatusCode: 401,
		},
	}

	client := privateMobileClient

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(`{"name": "Private Device Name", "pushToken": "token"}`).
				SetHeaders(c.headers).
				Post("device")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}
		})
	}
}

func TestPublicDevicePasswordChange(t *testing.T) {
	device := mobileDeviceRegister(t, publicMobileClient)

	cases := []struct {
		name               string
		headers            map[string]string
		body               string
		expectedStatusCode int
	}{
		{
			name: "with invalid token",
			headers: map[string]string{
				"Authorization": "Bearer 123456789",
			},
			body:               `{"currentPassword": "123456789", "newPassword": "123456789"}`,
			expectedStatusCode: 401,
		},
		{
			name: "with invalid password",
			headers: map[string]string{
				"Authorization": "Bearer " + device.Token,
			},
			body:               `{"currentPassword": "123456789", "newPassword": "changemeonemoretime"}`,
			expectedStatusCode: 401,
		},
		{
			name: "short password",
			headers: map[string]string{
				"Authorization": "Bearer " + device.Token,
			},
			body:               `{"currentPassword": "` + device.Password + `", "newPassword": "changeme"}`,
			expectedStatusCode: 400,
		},
		{
			name: "success",
			headers: map[string]string{
				"Authorization": "Bearer " + device.Token,
			},
			body:               `{"currentPassword": "` + device.Password + `", "newPassword": "changemeonemoretime"}`,
			expectedStatusCode: 204,
		},
		{
			name: "success with new password",
			headers: map[string]string{
				"Authorization": "Bearer " + device.Token,
			},
			body:               `{"currentPassword": "changemeonemoretime", "newPassword": "` + device.Password + `"}`,
			expectedStatusCode: 204,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := publicMobileClient.R().
				SetHeader("Content-Type", "application/json").
				SetBody(c.body).
				SetHeaders(c.headers).
				Patch("user/password")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}
		})
	}
}

func TestPublicDeviceRegisterWithCredentials(t *testing.T) {
	// won't work with registration rate limits
	t.SkipNow()

	firstDevice := mobileDeviceRegister(t, publicMobileClient)

	cases := []struct {
		name               string
		headers            map[string]string
		expectedStatusCode int
		expectedLogin      string
		expectedPassword   string
	}{
		{
			name: "Valid Credentials",
			headers: map[string]string{
				"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(firstDevice.Login+":"+firstDevice.Password)),
			},
			expectedStatusCode: 201,
			expectedLogin:      "",
			expectedPassword:   "",
		},
		{
			name: "Invalid Credentials",
			headers: map[string]string{
				"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("a:1")),
			},
			expectedStatusCode: 401,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := publicMobileClient.R().
				SetHeader("Content-Type", "application/json").
				SetBody(`{"name": "Public Device Name", "pushToken": "token"}`).
				SetHeaders(c.headers).
				Post("device")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}

			if !res.IsSuccess() {
				return
			}

			var resp mobileRegisterResponse
			if err := json.Unmarshal(res.Body(), &resp); err != nil {
				t.Fatal(err)
			}

			if resp.Login != c.expectedLogin {
				t.Fatalf("expected login %s, got %s", c.expectedLogin, resp.Login)
			}

			if resp.Password != c.expectedPassword {
				t.Fatalf("expected password %s, got %s", c.expectedPassword, resp.Password)
			}
		})
	}
}
