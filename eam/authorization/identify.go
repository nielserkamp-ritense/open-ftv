package authorization

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// Identify validates the caller of req once and returns the request-scoped Principal, so a
// handler can hand it to the management-plane services (ADR 0006). It performs no decision.
//
// A caller that fails authentication yields the error; a request without any identity
// yields the system principal, which the PDP then denies unless it runs fail-open.
func (a *auth) Identify(req *Request) (*RequestPrincipal, error) {
	parc := a.pep.PARCFromRequest(&models.Request{
		UID:     req.UID,
		URL:     req.URL,
		Method:  req.Method,
		Headers: req.Headers,
		Body:    req.Body,
	}, a.getter)

	if err := a.authenticate(parc); err != nil {
		return SystemRequestPrincipal(), err
	}

	return principalFrom(parc), nil
}

// authenticate checks the credentials the PEP collected, when an authenticator is configured.
func (a *auth) authenticate(parc *models.PARC) error {
	if a.authenticator == nil {
		return nil
	}

	user := convert.AnyToString(parc.Context.GetAttributeValue(models.AttrBasicUser))
	pswd := convert.AnyToString(parc.Context.GetAttributeValue(models.AttrBasicPswd))
	apikey := convert.AnyToString(parc.Context.GetAttributeValue(models.AttrAPIKey))

	if user == "" && apikey != "" {
		return a.authenticator.AuthenticateApiKey(a.ctx, apikey)
	}

	return a.authenticator.AuthenticateUser(a.ctx, user, pswd)
}

// principalFrom turns the principal the PEP derived into the request-scoped Principal.
//
// A user keeps its id and roles; an app (API key) keeps its id; anything else, including
// the PEP's "invalid" placeholder for a request without credentials, is the system.
func principalFrom(parc *models.PARC) *RequestPrincipal {
	attrs := parc.Principal.Attributes()

	p := identity.FromEntity(parc.Principal)
	p.Name = convert.AnyToString(attrs.GetAttributeValue(models.AttrPreferredName))
	p.Email = convert.AnyToString(attrs.GetAttributeValue(models.AttrEmail))
	p.Issuer = convert.AnyToString(attrs.GetAttributeValue(models.AttrIssuer))

	switch {
	case p.IsAuthenticatedUser():
		roles, _ := attrs.GetAttributeValue(models.AttrRoles).([]string)
		return &RequestPrincipal{Principal: p, Roles: roles}
	case p.Kind == identity.KindApp && p.ID != "":
		return &RequestPrincipal{Principal: p}
	default:
		return SystemRequestPrincipal()
	}
}
