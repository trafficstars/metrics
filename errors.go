package metrics

import (
	"errors"
)

var (
	// ErrAlreadyExists should never be returned: it's an internal error.
	// If you get this error then please let us know.
	ErrAlreadyExists = errors.New(`such metric is already registered`)
)
