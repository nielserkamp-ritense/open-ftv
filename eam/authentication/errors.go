package authentication

import "fmt"

// ErrUnauthenticated represents the error for unauthenticated access.
type ErrUnauthenticated struct {
	err error
}

// Unauthenticated wraps err as an ErrUnauthenticated.
func Unauthenticated(err error) *ErrUnauthenticated {
	return &ErrUnauthenticated{err: err}
}

// Error implements the error interface.
func (e *ErrUnauthenticated) Error() string {
	return fmt.Sprintf("authentication failure: %v", e.err)
}
