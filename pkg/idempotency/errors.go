package idempotency

import "errors"

var (
	ErrNotFound = errors.New("idempotency key not found")
)