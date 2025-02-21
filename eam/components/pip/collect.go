package pip

import (
	"log/slog"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// CollectAttributesFromRequest uses the given PBAC authorization request and other inputs
// to collect a set of attributes to be used by the Policy Decision Point.
//
// The default attributes stored in the PIP will be collected first.
// AttributeSet from the request will overwrite default attributes when the keys are equal.
func (p *pip) CollectAttributesFromRequest(req *models.Request) (models.AttributeSet, string) {
	a := p.newAttributes(p.attributes)
	a.AddAttribute(models.AttrRequestTime, time.Now().UTC())

	newURI := p.testHeaders(req, a)
	if newURI == "" && req.Resource != nil {
		newURI = req.Resource.ID()
	}

	p.determineURL(req, a)
	p.decodeBody(req, a)

	if req.Principal != nil {
		a.AddAttribute(models.AttrPrincipal, req.Principal.UID())
	}
	if req.Action != nil {
		a.AddAttribute(models.AttrAction, req.Action.UID())
	}
	if req.Resource != nil {
		a.AddAttribute(models.AttrResource, req.Resource.UID())
	}

	if len(req.Attributes) > 0 {
		for k := range req.Attributes {
			a.AddAttribute(k, req.Attributes[k])
		}
	}

	// copy important key attributes.
	for _, key := range []string{models.AttrClientIP, models.AttrRvvaID, models.AttrTraceParent, models.AttrTraceState} {
		if attr := a.GetAttribute(key); attr != nil {
			req.Attributes[key] = attr.Value()
		}
	}

	if p.logger.Enabled(nil, slog.LevelDebug) {
		kv := make(map[string]any)
		a.IterateAttributes(func(attr models.Attribute) {
			kv[attr.Key()] = attr.Value()
		})
		p.logger.Debug("attributes collected", "request-uid", req.UID, "attributes", kv)
	}

	return a, newURI
}
