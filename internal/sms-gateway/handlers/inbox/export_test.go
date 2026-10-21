package inbox

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/inbox"
)

// ToMessageInputs maps the upload request to the domain inputs.
func ToMessageInputs(req smsgateway.MobilePostInboxRequest) []inbox.MessageInput {
	return toMessageInputs(req)
}
