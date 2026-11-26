package push_test

import (
	"testing"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
)

func TestEventWrapperKeyIncludesTokenType(t *testing.T) {
	t.Parallel()

	legacy := push.NewEventWrapper("value", smsgateway.PushTokenTypeToken)
	fid := push.NewEventWrapper("value", smsgateway.PushTokenTypeFID)

	if push.EventWrapperKey(legacy) == push.EventWrapperKey(fid) {
		t.Errorf("keys collide across token types: %q", push.EventWrapperKey(legacy))
	}
}
