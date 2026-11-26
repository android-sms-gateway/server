package fcm_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/fcm"
)

func TestNewMessageTargetsFID(t *testing.T) {
	t.Parallel()

	message := fcm.NewMessage(map[string]string{"event": "MessageEnqueued"}, "device-fid", smsgateway.PushTokenTypeFID)

	body, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	if want := `"fid":"device-fid"`; !strings.Contains(string(body), want) {
		t.Errorf("FID target missing, want %s in %s", want, body)
	}
	if strings.Contains(string(body), `"token"`) {
		t.Errorf("legacy token target must be empty, got %s", body)
	}
}

func TestNewMessageTargetsLegacyToken(t *testing.T) {
	t.Parallel()

	message := fcm.NewMessage(
		map[string]string{"event": "MessageEnqueued"},
		"legacy-token",
		smsgateway.PushTokenTypeToken,
	)

	body, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	if want := `"token":"legacy-token"`; !strings.Contains(string(body), want) {
		t.Errorf("legacy token target missing, want %s in %s", want, body)
	}
	if strings.Contains(string(body), `"fid"`) {
		t.Errorf("FID target must be empty, got %s", body)
	}
}
