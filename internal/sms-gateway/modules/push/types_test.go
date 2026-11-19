package push_test

import (
	"testing"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/client"
)

func TestEventWrapperKeyIncludesTokenType(t *testing.T) {
	t.Parallel()

	legacy := push.NewEventWrapper("value", client.TokenTypeToken)
	fid := push.NewEventWrapper("value", client.TokenTypeFID)

	if push.EventWrapperKey(legacy) == push.EventWrapperKey(fid) {
		t.Errorf("keys collide across token types: %q", push.EventWrapperKey(legacy))
	}
}
