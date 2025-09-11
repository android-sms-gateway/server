package cache

import "errors"

var (
	// ErrKeyNotFound indicates no value exists for the given key.
	ErrKeyNotFound = errors.New("key not found")
	// ErrKeyExpired indicates a value exists but has expired.
	ErrKeyExpired = errors.New("key expired")
	// ErrKeyExists indicates a conflicting set when the key already exists.
	ErrKeyExists = errors.New("key already exists")
)
