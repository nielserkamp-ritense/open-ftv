package pip

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

func (p *pip) testHeaders(req *standards.Request, a standards.AttributeSet) string {
	other := make(map[string]string)

	var activityID, newURI, fwd1, fwd2 string

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
				activityID = list[0]
				a.AddAttribute(standards.AttrActivityID, activityID)
				p.convertActivityID(activityID, a)
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
			case "new-uri":
				newURI = list[0]
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

func (p *pip) convertActivityID(id string, a standards.AttributeSet) {
	if p.entities == nil {
		return
	}

	// An activity is an entity with type 'rva'.
	e := p.GetEntity(fmt.Sprintf("rva::%s", id))
	if e == nil {
		return
	}

	// It should give us the actual 'doelbinding' from its attributes.
	doel, ok := e.Attributes().GetAttribute(standards.AttrDoelbinding).(string)
	if !ok || doel == "" {
		return
	}

	a.AddAttribute(standards.AttrDoelbinding, doel)
}
