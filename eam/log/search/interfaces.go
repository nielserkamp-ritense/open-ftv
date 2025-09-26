package search

import (
	"context"

	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authlog"
)

// Searcher represents the interface for querying an Authorization Decision Log.
type Searcher interface {
	Search(context.Context, *Criteria) (oas.AuthlogEntries, error)
}

// ParameterError represents the error type for parameter errors.
type ParameterError struct {
	msg string
}

// Error implements the Error interface.
func (e *ParameterError) Error() string {
	return e.msg
}
