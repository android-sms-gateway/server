package push

import "github.com/android-sms-gateway/client-go/smsgateway"

func NewEventWrapper(token string, tokenType smsgateway.PushTokenType) *eventWrapper {
	return &eventWrapper{Token: token, TokenType: tokenType}
}

func EventWrapperKey(wrapper *eventWrapper) string {
	return wrapper.key()
}
