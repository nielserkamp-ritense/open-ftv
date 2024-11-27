package pip

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/types"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

func (p *pip) determineURL(req *types.Request, a types.AttributeSet) {
	m := make(map[string]any, 8)

	if req.Method != "" {
		m[standards.AttrMethod] = req.Method
	}

	if req.RequestTime != nil {
		m[standards.AttrRequestTime] = *req.RequestTime
	}

	if req.URL != nil {
		s, _ := strings.CutSuffix(req.URL.Scheme, "//")
		s, _ = strings.CutSuffix(s, ":")

		inQ, outQ := req.URL.Query(), make(map[string]string)
		for k := range inQ {
			outQ[k] = strings.Join(inQ[k], "\r")
		}

		path := req.URL.Path

		m[standards.AttrScheme] = s
		m[standards.AttrQuery] = outQ
		m[standards.AttrHost] = req.URL.Host
		m[standards.AttrPath] = path

		if path != "" {
			path = strings.Trim(path, "/")
			m[standards.AttrPathParts] = strings.Split(path, "/")
		}
	}

	a.AddAttribute(standards.AttrHttp, m)
}
