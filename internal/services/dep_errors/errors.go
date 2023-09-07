package dep_errors

import "errors"

var (
	ErrInvalidJson     = errors.New("invalid json")
	ErrPackageNotFound = errors.New("package not found")
)
