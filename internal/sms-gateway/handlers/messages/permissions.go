// Package messages defines permission scopes for message-related operations.
package messages

import "github.com/android-sms-gateway/client-go/smsgateway"

const (
	// ScopeSend is the permission scope required for sending messages.
	ScopeSend = smsgateway.ScopeMessagesSend
	// ScopeRead is the permission scope required for reading individual messages.
	ScopeRead = smsgateway.ScopeMessagesRead
	// ScopeList is the permission scope required for listing messages.
	ScopeList = smsgateway.ScopeMessagesList
	// ScopeCancel is the permission scope required for cancelling messages.
	ScopeCancel = smsgateway.ScopeMessagesCancel
	// ScopeExport is the permission scope required for exporting messages.
	ScopeExport = smsgateway.ScopeMessagesExport
)
