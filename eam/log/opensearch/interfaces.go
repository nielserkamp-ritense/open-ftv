// Package authlog contains functionality for logging authorization decisions.
package opensearch

import "context"

// Logger represents the interface for an authorisation log sink.
type Logger interface {
	Log(ctx context.Context, wait bool, record *AuthRecord) error
}
