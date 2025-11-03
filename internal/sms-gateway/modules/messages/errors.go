package messages

import "errors"

var (
	ErrLockFailed = errors.New("failed to acquire lock")
	ErrNoContent  = errors.New("no text or data content")
)

type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}
