package auth

import "time"

const codeTTL = 5 * time.Minute

type Mode string

const (
	ModePublic  Mode = "public"
	ModePrivate Mode = "private"
)

// AuthCode is a one-time user authorization code
type AuthCode struct {
	Code       string
	ValidUntil time.Time
}
