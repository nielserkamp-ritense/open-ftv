package authorization

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
)

// RequestPrincipal is the request-scoped Principal of the management plane (see CONTEXT.md):
// who is acting, plus the roles the PDP evaluates. Roles sit next to identity.Principal rather
// than inside it because they must never reach the stored Principal record (ADR 0004).
type RequestPrincipal struct {
	identity.Principal
	Roles []string
}

// Label renders the caller for logs. An app is identified by its API key, which is a secret,
// so an app is labeled by kind only.
func (p *RequestPrincipal) Label() string {
	if p.Kind == identity.KindApp {
		return identity.KindApp.String() + "::*"
	}

	return p.String()
}

// SystemRequestPrincipal is the RequestPrincipal for the manager's own actions, and the
// fallback when no caller was identified.
func SystemRequestPrincipal() *RequestPrincipal {
	return &RequestPrincipal{Principal: identity.NewSystemPrincipal()}
}

// Trace carries the W3C trace context of a request into the authorization context.
type Trace struct {
	Parent string // traceparent header.
	State  string // tracestate header.
}

type ctxKey int

const (
	ctxPrincipal ctxKey = iota + 1
	ctxTrace
)

// WithRequestPrincipal attaches the identified caller to ctx.
func WithRequestPrincipal(ctx context.Context, p *RequestPrincipal) context.Context {
	return context.WithValue(ctx, ctxPrincipal, p)
}

// RequestPrincipalFromContext retrieves the caller attached via WithRequestPrincipal, if any.
func RequestPrincipalFromContext(ctx context.Context) (*RequestPrincipal, bool) {
	p, ok := ctx.Value(ctxPrincipal).(*RequestPrincipal)
	return p, ok
}

// WithTrace attaches the request's trace context to ctx.
func WithTrace(ctx context.Context, t Trace) context.Context {
	return context.WithValue(ctx, ctxTrace, t)
}

// TraceFromContext retrieves the trace context attached via WithTrace; zero when absent.
func TraceFromContext(ctx context.Context) Trace {
	t, _ := ctx.Value(ctxTrace).(Trace)
	return t
}
