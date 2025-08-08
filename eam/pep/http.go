package pep

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func (c *collector) processHTTP() {
	m := make(map[string]any, 8)

	if c.req.Method != "" {
		m[models.AttrMethod] = c.req.Method
	}

	if c.req.RequestTime != nil {
		m[models.AttrRequestTime] = *c.req.RequestTime
	}

	if c.req.URL != nil {
		if c.newURI == "" {
			c.newURI = c.req.URL.String()
		}

		s, _ := strings.CutSuffix(c.req.URL.Scheme, "//")
		s, _ = strings.CutSuffix(s, ":")

		inQ, outQ := c.req.URL.Query(), make(map[string]string)
		for k := range inQ {
			outQ[k] = strings.Join(inQ[k], "\r")
		}

		path := c.req.URL.Path

		m[models.AttrScheme] = s
		m[models.AttrQuery] = outQ
		m[models.AttrHost] = c.req.URL.Host
		m[models.AttrPath] = path

		if path != "" {
			path = strings.Trim(path, "/")
			m[models.AttrPathParts] = strings.Split(path, "/")
		}
	}

	c.parc.Context.AddAttributeKV(models.AttrHTTP, m)
}
