package authorization

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

// Authorizer represents the interface for authorizing local API requests.
// For instance, to protect a PAP or PIP against unauthorized access of their CRUD endpoints.
type Authorizer interface {
	Authorize(req *Request) (*models.Response, identity.Principal, error)
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

	// Resource optionally carries the stored target object (with its attributes, e.g. status) so the PDP can evaluate
	// fine-grained, resource-attribute policies. When nil, the PEP synthesizes one.
	Resource *models.Entity
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

	if req.Resource != nil {
		r.Resource = req.Resource
	}

	parc := a.pep.PARCFromRequest(r, a.getter)

	if a.authenticator != nil {
		user := convert.AnyToString(parc.Context.GetAttributeValue(models.AttrBasicUser))
		pswd := convert.AnyToString(parc.Context.GetAttributeValue(models.AttrBasicPswd))
		apikey := convert.AnyToString(parc.Context.GetAttributeValue(models.AttrAPIKey))

		switch {
		case user == "" && apikey != "":
			err = a.authenticator.AuthenticateApiKey(a.ctx, apikey)
		default:
			err = a.authenticator.AuthenticateUser(a.ctx, user, pswd)
		}

		if err != nil {
			return
		}
	}

	// Reuse parc.Principal, already resolved by eam/pep's DeterminePrincipal.
	// Non-user kinds fall back to the system sentinel below.
	principal = identity.FromEntity(parc.Principal)
	principal.Name = convert.AnyToString(parc.Principal.Attributes().GetAttributeValue(models.AttrPreferredName))

	if !principal.IsAuthenticatedUser() {
		principal = identity.NewSystemPrincipal()
	}

	if a.noAuth {
		resp = &models.Response{Allowed: true}
	} else {
		resp, err = a.pdp.Authorize(req.UID.String(), parc)
	}

	return
}

type auth struct {
	ctx           context.Context
	log           *slog.Logger
	pep           *pep.PEP
	pdp           pdp.Controller
	getter        models.GetEntity
	authenticator authentication.Authenticator
	noAuth        bool
	debug         bool
}

func dummyGetter(string) (*models.Entity, uint64, error) { return nil, 0, nil }
