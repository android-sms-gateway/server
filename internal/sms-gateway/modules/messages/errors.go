package messages

import "errors"

var (
	ErrMessageAlreadyExists  = errors.New("duplicate id")
	ErrMessageNotFound       = errors.New("message not found")
	ErrMultipleMessagesFound = errors.New("multiple messages found")
	ErrNoContent             = errors.New("no text or data content")

	ErrQueueLimitExceeded = errors.New("queue limits exceeded")
)

type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}
