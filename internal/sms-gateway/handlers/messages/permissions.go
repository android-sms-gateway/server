// Package messages defines permission scopes for message-related operations.
package messages

const (
	// ScopeSend is the permission scope required for sending messages.
	ScopeSend = "messages:send"
	// ScopeRead is the permission scope required for reading individual messages.
	ScopeRead = "messages:read"
	// ScopeList is the permission scope required for listing messages.
	ScopeList = "messages:list"
	// ScopeExport is the permission scope required for exporting messages.
	ScopeExport = "messages:export"
)
