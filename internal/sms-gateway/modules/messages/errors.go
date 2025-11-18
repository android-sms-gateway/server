package messages

import "errors"

var (
	ErrNoContent = errors.New("no text or data content")
)

type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}
