package users

import "errors"

var (
	ErrNotFound = errors.New("user not found")
	ErrExists   = errors.New("user already exists")
)
