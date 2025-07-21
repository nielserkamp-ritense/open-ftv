package pep

import (
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func (c *collector) run() {
	parc, attrs := c.parc, c.parc.Context

	// all objects must be instantiated before processing.
	if parc.Principal == nil {
		parc.Principal = models.NewEntity("", "", models.NewAttributeSet())
	}
	if parc.Action == nil {
		parc.Action = models.NewEntity("", "", models.NewAttributeSet())
	}
	if parc.Resource == nil {
		parc.Resource = models.NewEntity("", "", models.NewAttributeSet())
	}

	if a := attrs.GetAttribute(models.AttrTime); a == nil {
		// See AuthZEN spec Information Model - Context (the link is subject to change):
		// https://openid.net/specs/authorization-api-1_0-01.html#name-context
		attrs.AddAttribute(models.AttrTime, time.Now().UTC())
	}

	// process the HTTP request data.
	c.processHeaders()
	if c.newURI == "" {
		c.newURI = parc.Resource.ID()
	}
	c.processHTTP()
	c.decodeBody()

	// let's make sure our principal, action & resource are properly filled.
	c.determinePrincipal()
	c.determineAction()
	c.determineResource()

	// TODO: are these duplications really needed?
	if p := parc.Principal; p.ID() != "" && p.ID() != PrincipalInvalid {
		attrs.AddAttribute(models.AttrPrincipal, parc.Principal.UID())
	}
	if parc.Action.ID() != "" {
		attrs.AddAttribute(models.AttrAction, parc.Action.UID())
	}
	if parc.Resource.ID() != "" {
		attrs.AddAttribute(models.AttrResource, parc.Resource.UID())
	}

	if c.debug {
		kv := make(map[string]any)
		attrs.IterateAttributes(func(attr *models.Attribute) {
			kv[attr.Key()] = attr.Value()
		})
		c.logger.Debug("attributes collected", "attributes", kv)
	}
}
