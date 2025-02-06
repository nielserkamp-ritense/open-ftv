package pip

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

func (p *pip) testHeaders(req *components.Request, a models.AttributeSet) string {
	other := make(map[string]string)

	var activityID, newURI, fwd1, fwd2 string

	for k := range req.Headers {
		if list := req.Headers[k]; len(list) > 0 {
			switch strings.ToLower(k) {
			case models.HeaderContentType:
				a.AddAttribute(models.AttrContentType, list[0])
			case models.HeaderAuthorization:
				p.processAuth(req, list[0], a)
			case models.HeaderFSCAuthorization:
				newURI = p.processFSC(req, list[0], a)
			case models.HeaderApiKey, "apikey", "x-apikey", "x-api-key":
				a.AddAttribute(models.AttrApiKey, list[0])
			case models.HeaderRvvaID, models.HeaderObsoleteRvvaID:
				activityID = list[0]
				a.AddAttribute(models.AttrRvvaID, activityID)
				p.convertActivityID(activityID, a)
			case models.HeaderCoreUser, models.HeaderObsoleteCoreUser:
				a.AddAttribute(models.AttrCoreUser, list[0])
			case models.HeaderGrondslag:
				a.AddAttribute(models.AttrGrondslag, list[0])
			case models.HeaderDoelbinding:
				a.AddAttribute(models.AttrDoelbinding, list[0])
			case models.HeaderZaakType, "zaaktype":
				a.AddAttribute(models.AttrZaakType, list[0])
			case models.HeaderTaak:
				a.AddAttribute(models.AttrTaak, list[0])
			case models.HeaderXForwardedFor:
				fwd1 = strings.Join(list, ",")
			case models.HeaderForwarded:
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

	a.AddAttribute(models.AttrHeaders, other)

	return newURI
}

func (p *pip) convertActivityID(id string, a models.AttributeSet) {
	if p.entities == nil {
		return
	}

	// An activity is an entity with type 'rva'.
	e := p.GetEntity(fmt.Sprintf("rva::%s", id))
	if e == nil {
		return
	}

	// It should give us the actual 'doelbinding' from its attributes.
	doel, ok := e.Attributes().GetAttributeValue(models.AttrDoelbinding).(string)
	if !ok || doel == "" {
		return
	}

	a.AddAttribute(models.AttrDoelbinding, doel)
}
