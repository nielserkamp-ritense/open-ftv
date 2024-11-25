package pip

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/types"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

func (p *pip) processHeaders(req *types.Request, a types.AttributeSet) string {
	other := make(map[string]string)

	var newURI, fwd1, fwd2 string

	for k := range req.Headers {
		if list := req.Headers[k]; len(list) > 0 {
			switch strings.ToLower(k) {
			case standards.HeaderContentType:
				a.AddAttribute(standards.AttrContentType, list[0])
			case standards.HeaderAuthorization:
				p.processAuth(req, list[0], a)
			case standards.HeaderFSCAuthorization:
				newURI = p.processFSC(req, list[0], a)
			case standards.HeaderApiKey, "apikey", "x-apikey", "x-api-key":
				a.AddAttribute(standards.AttrApiKey, list[0])
			case standards.HeaderRvaActivityID:
				a.AddAttribute(standards.AttrActivityID, list[0])
			case standards.HeaderCoreUser:
				a.AddAttribute(standards.AttrCoreUser, list[0])
			case standards.HeaderGrondslag:
				a.AddAttribute(standards.AttrGrondslag, list[0])
			case standards.HeaderDoelbinding:
				a.AddAttribute(standards.AttrDoelbinding, list[0])
			case standards.HeaderZaakType, "zaaktype":
				a.AddAttribute(standards.AttrZaakType, list[0])
			case standards.HeaderTaak:
				a.AddAttribute(standards.AttrTaak, list[0])
			case standards.HeaderXForwardedFor:
				fwd1 = strings.Join(list, ",")
			case standards.HeaderForwarded:
				fwd2 = strings.Join(list, ",")
			default:
				other[k] = strings.Join(list, ",")
			}
		}
	}

	if fwd1 != "" || fwd2 != "" {
		p.processForwarded(fwd1, fwd2, a)
	}

	a.AddAttribute(standards.AttrHeaders, other)

	return newURI
}
