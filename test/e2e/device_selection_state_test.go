package e2e

import (
	"encoding/json"
	"testing"
)

func TestDeviceSelectionDisabled(t *testing.T) {
	device := mobileDeviceRegister(t, publicMobileClient)
	client := publicUserClient.Clone().SetBasicAuth(device.Login, device.Password)

	disabled := mobileDeviceRegister(
		t,
		publicMobileClient,
		(&mobileDeviceRegisterOptions{}).withCredentials(device.Login, device.Password),
	)

	patch := func(state string) {
		res, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(map[string]any{"state": state}).
			Patch("devices/" + disabled.ID)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode() != 204 {
			t.Fatal(res.StatusCode(), res.String())
		}
	}

	post := func(deviceID string, expectedStatus int) messageState {
		body := map[string]any{
			"textMessage":  map[string]any{"text": "test"},
			"phoneNumbers": []string{"+79999999999"},
		}
		if deviceID != "" {
			body["deviceId"] = deviceID
		}

		res, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Post("messages")
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode() != expectedStatus {
			t.Fatal(res.StatusCode(), res.String())
		}
		if expectedStatus != 202 {
			return messageState{}
		}

		var state messageState
		if err := json.Unmarshal(res.Body(), &state); err != nil {
			t.Fatal(err)
		}

		return state
	}

	patch("Disabled")

	if post(disabled.ID, 400).ID != "" {
		t.Fatal("explicit send to a disabled device must fail")
	}

	for range 5 {
		if state := post("", 202); state.DeviceID == disabled.ID {
			t.Fatalf("random selection picked disabled device %s", disabled.ID)
		}
	}

	patch("Active")

	if state := post(disabled.ID, 202); state.DeviceID != disabled.ID {
		t.Fatalf("expected re-enabled device %s, got %s", disabled.ID, state.DeviceID)
	}
}
