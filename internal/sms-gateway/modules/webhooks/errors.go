package webhooks

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidEvent = errors.New("invalid event")
)

type ValidationError struct {
	Field string
	Value string
	Err   error
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("invalid %q = %q: %s", e.Field, e.Value, e.Err)
}

func (e ValidationError) Unwrap() error {
	return e.Err
}

func newValidationError(field, value string, err error) ValidationError {
	return ValidationError{
		Field: field,
		Value: value,
		Err:   err,
	}
}

func IsValidationError(err error) bool {
	return errors.As(err, new(ValidationError))
}
