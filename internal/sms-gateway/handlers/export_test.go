package handlers

import "github.com/android-sms-gateway/client-go/smsgateway"

func PushTokenTypeToString(tokenType *smsgateway.PushTokenType) string {
	return pushTokenTypeToString(tokenType)
}
