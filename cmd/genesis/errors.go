// Path: genesis/cmd/genesis/errors.go
package main

import "errors"

var (
	// ErrDeterminantMissing maps to Exit Code 2
	ErrDeterminantMissing = errors.New("determinant missing or malformed")
	
	// ErrAccessDenied maps to Exit Code 126
	ErrAccessDenied = errors.New("access denied")
	
	// ErrBoundaryViolation maps to Exit Code 1
	ErrBoundaryViolation = errors.New("boundary law violation")
)
