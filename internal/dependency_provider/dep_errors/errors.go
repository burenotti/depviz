package dep_errors

import (
	"errors"
)

var (
	ErrFetch           = errors.New("invalid json")
	ErrPackageNotFound = errors.New("package not found")
)
