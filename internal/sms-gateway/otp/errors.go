package otp

import "errors"

var (
	// ErrInvalidConfig indicates that the OTP configuration is invalid.
	ErrInvalidConfig = errors.New("invalid config")
	// ErrInitFailed indicates that the OTP service failed to initialize.
	ErrInitFailed = errors.New("initialization failed")
)
