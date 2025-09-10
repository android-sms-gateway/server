package cache

import "errors"

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
	ErrKeyExists   = errors.New("key already exists")
)
