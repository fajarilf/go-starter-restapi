package domain

import "errors"

// ErrNotFound is returned by repositories/services when a requested resource
// does not exist. Handlers map it to HTTP 404.
var ErrNotFound = errors.New("not found")
