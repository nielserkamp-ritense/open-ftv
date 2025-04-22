package pep

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"

func (c *collector) determineAction() {
	if c.req.Method != "" {
		var a string
		switch c.req.Method {
		case "PUT":
			a = "can_create"
		case "POST":
			a = "can_update"
		case "DELETE":
			a = "can_delete"
		default:
			// TODO: perhaps other HTTP methods may require a different action id.
			a = "can_read"
		}

		if action := c.parc.Action; action.Type() == "" && action.ID() == "" {
			c.parc.Action = models.NewEntity(models.EntityTypeName, a, models.NewAttributeSet(c.parc.Action.Attributes()))
		}

		c.parc.Action.Attributes().AddAttribute(models.AttrMethod, c.req.Method)
	}
}
