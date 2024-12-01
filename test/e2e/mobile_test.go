package e2e

import (
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

func makeClient(baseUrl string) *resty.Client {
	return resty.New().
		SetBaseURL(baseUrl).
		SetTimeout(300 * time.Millisecond)
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

	client := makeClient(PublicURL + "/mobile/v1/device")

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(`{"name": "Public Device Name", "pushToken": "token"}`).
				SetHeaders(c.headers).
				Post("")
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

	client := makeClient(PrivateURL + "/mobile/v1/device")

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(`{"name": "Private Device Name", "pushToken": "token"}`).
				SetHeaders(c.headers).
				Post("")
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode() != c.expectedStatusCode {
				t.Fatal(res.StatusCode(), res.String())
			}
		})
	}
}
