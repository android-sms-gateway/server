package otp

import "time"

// Code is a one-time user authorization code.
type Code struct {
	Code       string
	ValidUntil time.Time
}
