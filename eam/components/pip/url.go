package pip

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

func (p *pip) determineURL(req *models.Request, a models.AttributeSet) {
	m := make(map[string]any, 8)

	if req.Method != "" {
		m[models.AttrMethod] = req.Method
	}

	if req.RequestTime != nil {
		m[models.AttrRequestTime] = *req.RequestTime
	}

	if req.URL != nil {
		s, _ := strings.CutSuffix(req.URL.Scheme, "//")
		s, _ = strings.CutSuffix(s, ":")

		inQ, outQ := req.URL.Query(), make(map[string]string)
		for k := range inQ {
			outQ[k] = strings.Join(inQ[k], "\r")
		}

		path := req.URL.Path

		m[models.AttrScheme] = s
		m[models.AttrQuery] = outQ
		m[models.AttrHost] = req.URL.Host
		m[models.AttrPath] = path

		if path != "" {
			path = strings.Trim(path, "/")
			m[models.AttrPathParts] = strings.Split(path, "/")
		}
	}

	a.AddAttribute(models.AttrHTTP, m)
}
