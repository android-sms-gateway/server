package e2e

import (
	"encoding/json"
	"testing"

	"github.com/go-resty/resty/v2"
)

type mobileDeviceRegisterOptions struct {
	username string
	password string
}

func (o *mobileDeviceRegisterOptions) withCredentials(username, password string) *mobileDeviceRegisterOptions {
	o.username = username
	o.password = password
	return o
}

func mobileDeviceRegister(t *testing.T, client *resty.Client, opts ...*mobileDeviceRegisterOptions) mobileRegisterResponse {
	req := client.R()
	for _, opt := range opts {
		if opt.username != "" && opt.password != "" {
			req.SetBasicAuth(opt.username, opt.password)
		}
	}

	res, err := req.
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
