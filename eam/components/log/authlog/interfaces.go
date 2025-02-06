package authlog

import "context"

// Logger represents the interface for an authorisation log sink.
type Logger interface {
	Log(ctx context.Context, record *AuthRecord) error
}
