package authorization

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

// Authorizer represents the interface for authorizing local API requests.
// For instance, to protect a PAP or PIP against unauthorized access of their CRUD endpoints.
type Authorizer interface {
	Authorize(req *Request) (*models.Response, error)
}

// New instantiates a new authorization handler.
func New(opts ...Option) Authorizer {
	// initialize with default context, logger and empty entity set.
	a := &auth{
		ctx:      context.Background(),
		log:      slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		entities: models.NewEntitySet(),
	}

	for i := range opts {
		opts[i](a)
	}

	// if a pep was not given, we'll create our own.
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
func (a *auth) Authorize(req *Request) (*models.Response, error) {
	r := &models.Request{
		UID:     req.UID,
		URL:     req.URL,
		Method:  req.Method,
		Headers: req.Headers,
		Body:    req.Body,
	}

	parc := a.pep.PARCFromRequest(r, a.entities)

	if a.authenticator != nil {
		user := convert.AnyToString(parc.Context.GetAttributeValue(models.AttrBasicUser))
		pswd := convert.AnyToString(parc.Context.GetAttributeValue(models.AttrBasicPswd))
		apikey := convert.AnyToString(parc.Context.GetAttributeValue(models.AttrAPIKey))

		var err error
		switch {
		case user != "":
			err = a.authenticator.AuthenticateUser(a.ctx, user, pswd)
		case apikey != "":
			err = a.authenticator.AuthenticateApiKey(a.ctx, apikey)
		default:
			err = &authentication.ErrUnauthenticated{}
		}

		if err != nil {
			return nil, err
		}
	}

	return a.pdp.Authorize(req.UID.String(), parc)
}

type auth struct {
	ctx           context.Context
	log           *slog.Logger
	pep           pep.PEP
	pdp           pdp.Controller
	entities      models.EntitySet
	authenticator authentication.Authenticator
	debug         bool
}
