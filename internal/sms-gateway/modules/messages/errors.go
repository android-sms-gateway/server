package messages

import "errors"

var (
	ErrMessageAlreadyExists  = errors.New("message with the same ID already exists")
	ErrMessageNotFound       = errors.New("message not found")
	ErrMultipleMessagesFound = errors.New("multiple messages found")
	ErrNoContent             = errors.New("no text, data or multimedia content")
	ErrMessageNotPending     = errors.New("message is not pending")
	ErrStaleCacheEntry       = errors.New("stale cache entry")

	ErrQueueLimitExceeded = errors.New("queue limits exceeded")
)

type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}
