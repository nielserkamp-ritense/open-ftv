// Package body provides the BodyToContext request mapper for OpenFTV.
//
// It detects a request body in the action attributes, decodes it, and exposes it
// as the "body" context attribute so that policies can match on its parsed
// contents. Importing this package registers the mapper under the configuration
// name "bodytocontext".
//
// Hardening: when a body is present but cannot be parsed — because it is not
// decodable, or because the Content-Type is missing/unusable and the content
// cannot be sniffed — the mapper does NOT silently continue. Instead it omits the
// "body" context attribute and sets an explicit "body_error" context attribute
// describing the reason. Policies that match on the body MUST therefore be written
// default-deny (see README): an absent "body" attribute must never be treated as
// an allow.
package body

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/decode"
)

// AttrBodyError is the context attribute key set when a present body could not be
// parsed into the "body" context attribute.
const AttrBodyError = "body_error"

// BodyToContext detects a body in the action attributes,
// and decodes it as a context attribute if possible.
//
// If no body is present (or it is empty), the given PARC is returned unmodified.
// If a body is present but cannot be parsed, the "body" attribute is omitted and
// the "body_error" context attribute is set with the reason instead.
func BodyToContext(parc *models.PARC, opts ...mapping.Option) *models.PARC {
	attr := parc.Action.Attributes().GetAttribute(models.AttrBody)
	if attr == nil {
		return parc
	}

	body, ok := attr.Value().(string)
	if !ok || body == "" {
		return parc
	}

	// decode the body into a byte array.
	d, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		unquoted, _ := strconv.Unquote(fmt.Sprintf(`"%s"`, body))
		d = []byte(unquoted)
	}

	// extract the content-type header.
	ct := mapping.FromHeaders(parc.Context.GetAttribute(models.AttrHeaders), []string{models.HeaderContentType}, opts...)

	m, err := decode.ParseBody(d, ct)
	if err != nil {
		reason := fmt.Sprintf("body present but could not be parsed (content-type %q): %v", ct, err)
		return withContext(parc, models.NewAttribute(AttrBodyError, reason))
	}

	return withContext(parc, models.NewAttribute(models.AttrBody, m))
}

// withContext returns a copy of parc with the given attribute added to the context.
func withContext(parc *models.PARC, attr models.Attribute) *models.PARC {
	return &models.PARC{
		Principal: parc.Principal,
		Action:    parc.Action,
		Resource:  parc.Resource,
		Context:   models.NewAttributeSet(parc.Context, attr),
	}
}
