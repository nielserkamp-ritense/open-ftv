package authorization

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
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

	if err := rejectedBearer(parc); err != nil {
		return SystemRequestPrincipal(), err
	}

	if err := a.authenticate(parc); err != nil {
		return SystemRequestPrincipal(), err
	}

	caller := principalFrom(parc)
	if caller.IsAuthenticatedUser() && len(caller.Roles) == 0 {
		caller.Roles = a.storedRoles(caller.ID)
	}

	return caller, nil
}

// storedRoles reads the roles of a user the entity store holds. A bearer token carries its
// roles as a claim; a basic-auth user is only a name and a password hash, and its roles live
// on the stored entity (the PIP), the way the PDP would resolve them.
func (a *auth) storedRoles(id string) []string {
	e, _, err := a.getter(identity.KindUser.String() + "::" + id)
	if err != nil || e == nil {
		return nil
	}

	return rolesOf(e.Attributes().GetAttributeValue(models.AttrRoles))
}

// rolesOf reads a roles attribute, which is a set for token users and may be one string
// for locally seeded users.
func rolesOf(v any) []string {
	switch roles := v.(type) {
	case []string:
		return roles
	case string:
		if roles == "" {
			return nil
		}

		return []string{roles}
	case []any:
		out := make([]string, 0, len(roles))

		for i := range roles {
			if s, ok := roles[i].(string); ok && s != "" {
				out = append(out, s)
			}
		}

		return out
	default:
		return nil
	}
}

// rejectedBearer reports a bearer token the PEP refused. A refused token is a failed
// authentication, never a request without credentials, so it is answered before the PDP.
func rejectedBearer(parc *models.PARC) error {
	jwt, _ := parc.Context.GetAttributeValue(models.AttrJWT).(map[string]any)
	if jwt == nil {
		return nil
	}

	if valid, _ := jwt[models.AttrValid].(bool); valid {
		return nil
	}

	return authentication.Unauthenticated(fmt.Errorf("bearer token rejected: %v", jwt[pep.AttrError]))
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
		return &RequestPrincipal{Principal: p, Roles: rolesOf(attrs.GetAttributeValue(models.AttrRoles))}
	case p.Kind == identity.KindApp && p.ID != "":
		return &RequestPrincipal{Principal: p}
	default:
		return SystemRequestPrincipal()
	}
}
