package authorization

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

// ErrPrincipalNotRecorded reports that the caller could not be recorded in the principal store.
//
// It is deliberately its own type: created_by/updated_by are foreign keys to principal, so the
// write that follows would fail with a foreign-key violation, and that is a server-side storage
// fault rather than a permission denial. Handlers map it to 500, never 403.
type ErrPrincipalNotRecorded struct{ Err error }

func (e *ErrPrincipalNotRecorded) Error() string {
	return "failed to record principal: " + e.Err.Error()
}

func (e *ErrPrincipalNotRecorded) Unwrap() error { return e.Err }

// PrincipalRecorder records that a Principal was seen acting on the management plane.
// Implemented by eam/principals; declared here so authorization keeps no storage dependency.
type PrincipalRecorder interface {
	Upsert(ctx context.Context, p *identity.Principal) error
}

// Authorizer represents the interface for authorizing local API requests.
// For instance, to protect a PAP or PIP against unauthorized access of their CRUD endpoints.
type Authorizer interface {
	// Authorize decides an HTTP request as a whole: the PEP derives action and resource from
	// the request. This is the request-path shape; the management plane moves to Identify and
	// Decide, handler by handler (ADR 0006).
	Authorize(req *Request) (*models.Response, identity.Principal, error)
	// Identify validates the caller once, without deciding anything.
	Identify(req *Request) (*RequestPrincipal, error)
	// Decide answers the management plane's question for an identified caller.
	Decide(ctx context.Context, caller *RequestPrincipal, action string, resource *models.Entity) error
}

// New instantiates a new authorization handler.
func New(opts ...Option) Authorizer {
	// initialize with default context, logger and empty entity set.
	a := &auth{
		ctx:    context.Background(),
		log:    slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		getter: dummyGetter,
	}

	for i := range opts {
		opts[i](a)
	}

	// if the pep was not given, we'll create our own.
	if a.pep == nil {
		a.pep = pep.New(a.ctx, a.log)
	}

	a.debug = a.log.Enabled(a.ctx, slog.LevelDebug)
	return a
}

// Request represents an HTTP request that must be authorized.
type Request struct {
	UID     *uuid.UUID
	URL     *url.URL
	Method  string
	Headers map[string][]string
	Body    []byte
}

// Authorize implements the Authorizer interface.
func (a *auth) Authorize(req *Request) (resp *models.Response, principal identity.Principal, err error) {
	principal = identity.NewUnknownPrincipal()

	r := &models.Request{
		UID:     req.UID,
		URL:     req.URL,
		Method:  req.Method,
		Headers: req.Headers,
		Body:    req.Body,
	}

	parc := a.pep.PARCFromRequest(r, a.getter)

	if err := a.authenticate(parc); err != nil {
		return resp, principal, err
	}

	// Reuse parc.Principal, already resolved by eam/pep's DeterminePrincipal.
	// Non-user kinds fall back to the system sentinel: attribution names users only.
	principal = principalFrom(parc).Principal
	if !principal.IsAuthenticatedUser() {
		principal = identity.NewSystemPrincipal()
	}

	if a.noAuth {
		resp = &models.Response{Allowed: true}
	} else {
		resp, err = a.pdp.Authorize(req.UID.String(), parc)
	}

	// Record the caller only once the request is permitted, so a token that is valid but allowed
	// nothing leaves no personal data behind. See docs/adr/0004.
	if err == nil && resp != nil && resp.Allowed {
		err = a.recordPrincipal(isSafeMethod(req.Method), &principal)
	}

	return resp, principal, err
}

// recordPrincipal stores the caller in the principal store, so actions attributed to their id can
// later be rendered as a person.
//
// A failure is fatal for a write and ignored for a read (safe). created_by/updated_by are
// foreign keys to principal, so a write whose principal row is missing would fail deep inside the
// handler; failing here instead turns that into an early, comprehensible error. A read has no such
// dependency, and a degraded database must never lock everyone out of the management UI.
func (a *auth) recordPrincipal(safe bool, p *identity.Principal) error {
	if a.principals == nil || !p.IsAuthenticatedUser() {
		return nil
	}

	err := a.principals.Upsert(a.ctx, p)
	if err == nil {
		return nil
	}

	if safe {
		a.log.Error("failed to record principal", "principal", p.String(), "error", err)
		return nil
	}

	return &ErrPrincipalNotRecorded{Err: err}
}

// isSafeMethod reports whether m cannot write, and therefore cannot depend on a principal row.
func isSafeMethod(m string) bool {
	switch strings.ToUpper(m) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

type auth struct {
	ctx           context.Context
	log           *slog.Logger
	pep           *pep.PEP
	pdp           pdp.Controller
	getter        models.GetEntity
	authenticator authentication.Authenticator
	principals    PrincipalRecorder
	noAuth        bool
	debug         bool
}

func dummyGetter(string) (*models.Entity, uint64, error) { return nil, 0, nil }
