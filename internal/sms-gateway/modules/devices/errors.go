package devices

import "errors"

var (
	ErrInvalidUser     = errors.New("invalid user")
	ErrInconsistentE2E = errors.New("public key and key version must be both provided or both absent")
)
