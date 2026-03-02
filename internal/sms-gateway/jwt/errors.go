package jwt

import "errors"

var (
	ErrDisabled        = errors.New("jwt disabled")
	ErrInitFailed      = errors.New("failed to initialize jwt")
	ErrInvalidConfig   = errors.New("invalid config")
	ErrInvalidParams   = errors.New("invalid params")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidTokenUse = errors.New("invalid token use")
	ErrTokenRevoked    = errors.New("token revoked")
	ErrTokenReplay     = errors.New("token replay detected")
)
