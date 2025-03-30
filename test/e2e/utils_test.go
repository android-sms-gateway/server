package e2e

import (
	"encoding/json"
	"testing"

	"github.com/go-resty/resty/v2"
)

func mobileDeviceRegister(t *testing.T, client *resty.Client) mobileRegisterResponse {
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"name": "Public Device Name", "pushToken": "token"}`).
		Post("device")
	if err != nil {
		t.Fatal(err)
	}

	if !res.IsSuccess() {
		t.Fatal(res.StatusCode(), res.String())
	}

	var resp mobileRegisterResponse
	if err := json.Unmarshal(res.Body(), &resp); err != nil {
		t.Fatal(err)
	}

	return resp
}
