package pep

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/decode"
)

func (c *collector) decodeBody() {
	if len(c.req.Body) == 0 {
		return
	}

	ct, _ := c.parc.Context.GetAttributeValue(models.AttrContentType).(string)
	attr, err := decode.ParseBody(c.req.Body, ct)
	if err != nil {
		c.logger.Error("failed to parse body", "content-type", ct, "error", err)
	} else {
		c.parc.Context.AddAttributeKV(models.AttrBody, attr)
	}
}
