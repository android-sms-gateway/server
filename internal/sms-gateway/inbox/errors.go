package inbox

import "errors"

var (
	ErrNotEncrypted = errors.New("inbox messages must be encrypted")
	ErrNotFound     = errors.New("inbox message not found")
	ErrEmptyBatch   = errors.New("batch must contain at least one message")
)
