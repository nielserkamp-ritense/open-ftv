package pep

import (
	"log/slog"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// PARCFromRequest uses the given authorization request and other inputs
// to collect the principal, action, resource and context to be used by the Policy Decision Point.
func (p *pep) PARCFromRequest(req *models.Request, e models.EntitySet) (*models.PARC, string) {
	logger := p.logger
	if req.UID != nil {
		logger = p.logger.With("uid", req.UID.String())
	}
	debug := logger.Enabled(nil, slog.LevelDebug)

	c := &collector{
		debug:  debug,
		logger: logger,
		req: &models.HTTPRequest{
			RequestTime: req.RequestTime,
			URL:         req.URL,
			Method:      req.Method,
			Headers:     req.Headers,
			Body:        req.Body,
		},
		parc: &models.PARC{
			Principal: req.Principal,
			Action:    req.Action,
			Resource:  req.Resource,
			Context:   models.NewAttributeSet(),
		},
		entities: e,
	}

	for k := range req.Attributes {
		v := req.Attributes[k]
		c.parc.Context.AddAttribute(k, v)
	}

	c.run()

	// TODO: move this into caller code; here is the wrong place for such hidden side-effects.
	// copy back important key attributes.
	for _, key := range []string{models.AttrClientIP, models.AttrRvvaID, models.AttrTraceParent, models.AttrTraceState} {
		if attr := c.parc.Context.GetAttribute(key); attr != nil {
			req.Attributes[key] = attr.Value()
		}
	}

	return c.parc, c.newURI
}

// PARCFromHTTP uses the given HTTP request and other inputs
// to collect the principal, action, resource and context to be used by the Policy Decision Point.
func (p *pep) PARCFromHTTP(uid uuid.UUID, req *models.HTTPRequest, attrs models.AttributeSet, e models.EntitySet) (*models.PARC, string) {
	logger := p.logger.With("uid", uid.String())
	debug := logger.Enabled(nil, slog.LevelDebug)

	c := &collector{
		debug:    debug,
		logger:   logger,
		req:      req,
		parc:     &models.PARC{Context: models.NewAttributeSet(attrs)},
		entities: e,
	}

	c.run()
	return c.parc, c.newURI
}

type collector struct {
	debug    bool
	logger   *slog.Logger
	req      *models.HTTPRequest
	parc     *models.PARC
	newURI   string
	entities models.EntitySet
}
