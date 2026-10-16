package devices

import "errors"

var (
	ErrNotFound      = errors.New("record not found")
	ErrInvalidFilter = errors.New("invalid filter")
	ErrMoreThanOne   = errors.New("more than one record")
)
