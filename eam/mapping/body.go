package mapping

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/decode"
)

// BodyToContext detects a body in the action attributes,
// and decodes it as a context attribute if possible.
//
// Any errors during decoding are ignored,
// in which case the body will not be present in the context.
func BodyToContext(parc *models.PARC, opts ...Option) *models.PARC {
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
		body, _ = strconv.Unquote(fmt.Sprintf(`"%s"`, body))
		d = []byte(body)
	}

	// extract the content-type header.
	b := &base{headerKeys: []string{models.HeaderContentType}}
	b.configure(opts)
	ct := b.fromHeaders(parc.Context.GetAttribute(models.AttrHeaders))

	var m map[string]any
	if m, err = decode.ParseBody(d, ct); err != nil {
		return parc
	}

	return &models.PARC{
		Principal: parc.Principal,
		Action:    parc.Action,
		Resource:  parc.Resource,
		Context:   models.NewAttributeSet(parc.Context, models.NewAttribute(models.AttrBody, m)),
	}
}
