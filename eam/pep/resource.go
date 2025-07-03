package pep

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"

func (c *collector) determineResource() {
	if r := c.parc.Resource; r.Type() != "" || r.ID() != "" {
		return
	}

	if c.newURI != "" {
		c.parc.Resource = models.NewEntity(models.EntityTypeService, c.newURI, models.NewAttributeSet(c.parc.Resource.Attributes()))
	}
}
