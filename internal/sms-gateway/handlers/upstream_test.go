package handlers_test

import (
	"testing"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers"
)

func TestPushTokenTypeToString(t *testing.T) {
	t.Parallel()

	if got := handlers.PushTokenTypeToString(nil); got != "" {
		t.Errorf("nil token type = %q, want empty", got)
	}

	fid := smsgateway.PushTokenTypeFID
	if got := handlers.PushTokenTypeToString(&fid); got != string(fid) {
		t.Errorf("FID token type = %q, want %q", got, fid)
	}

	legacy := smsgateway.PushTokenTypeToken
	if got := handlers.PushTokenTypeToString(&legacy); got != string(legacy) {
		t.Errorf("legacy token type = %q, want %q", got, legacy)
	}
}
