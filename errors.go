package metrics

import (
	"errors"
)

var (
	ErrAlreadyExists = errors.New(`Such metric is already registered`)
)
