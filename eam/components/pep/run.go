package pep

import (
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

func (c *collector) run() {
	parc, attrs := c.parc, c.parc.Context

	if parc.Principal == nil {
		parc.Principal = models.NewEntity("", "", models.NewAttributeSet())
	}
	if parc.Action == nil {
		parc.Action = models.NewEntity("", "", models.NewAttributeSet())
	}
	if parc.Resource == nil {
		parc.Resource = models.NewEntity("", "", models.NewAttributeSet())
	}

	attrs.AddAttribute(models.AttrRequestTime, time.Now().UTC())

	c.testHeaders()
	if c.newURI == "" && parc.Resource != nil {
		c.newURI = parc.Resource.ID()
	}

	c.determineURL()
	c.decodeBody()

	if parc.Principal.ID() != "" {
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
		attrs.IterateAttributes(func(attr models.Attribute) {
			kv[attr.Key()] = attr.Value()
		})
		c.logger.Debug("attributes collected", "attributes", kv)
	}
}
