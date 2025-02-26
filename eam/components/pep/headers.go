package pep

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

func (c *collector) processHeaders() {
	attrs := c.parc.Context
	other := make(map[string]string)

	var activityID, fwd1, fwd2 string

	for k := range c.req.Headers {
		if list := c.req.Headers[k]; len(list) > 0 {
			first := list[0]

			switch strings.ToLower(k) {
			case models.HeaderContentType:
				attrs.AddAttribute(models.AttrContentType, first)
			case models.HeaderAuthorization:
				c.processAuth(first)
			case models.HeaderFSCAuthorization:
				c.processFSC(first)
			case models.HeaderAPIKey, "apikey", "x-apikey", "x-api-key":
				attrs.AddAttribute(models.AttrAPIKey, first)
			case models.HeaderRvvaID, models.HeaderObsoleteRvvaID:
				activityID = first
				attrs.AddAttribute(models.AttrRvvaID, activityID)
				c.convertActivityID(activityID)
			case models.HeaderCoreUser, models.HeaderObsoleteCoreUser:
				attrs.AddAttribute(models.AttrCoreUser, first)
			case models.HeaderGrondslag:
				attrs.AddAttribute(models.AttrGrondslag, first)
			case models.HeaderDoelbinding:
				attrs.AddAttribute(models.AttrDoelbinding, first)
			case models.HeaderZaakType, "zaaktype":
				attrs.AddAttribute(models.AttrZaakType, first)
			case models.HeaderTaak:
				attrs.AddAttribute(models.AttrTaak, first)
			case models.HeaderXForwardedFor:
				fwd1 = strings.Join(list, ",")
			case models.HeaderForwarded:
				fwd2 = strings.Join(list, ",")
			case models.HeaderDeviceID, models.HeaderObsoleteDeviceID:
				c.parc.Principal.Attributes().AddAttribute(models.AttrDeviceID, first)
			case models.HeaderTraceParent:
				attrs.AddAttribute(models.AttrTraceParent, first)
			case models.HeaderTraceState:
				attrs.AddAttribute(models.AttrTraceState, first)
			case "new-uri":
				c.newURI = first
			default:
				other[k] = strings.Join(list, ",")
			}
		}
	}

	if fwd1 != "" || fwd2 != "" {
		c.processForwarded(fwd1, fwd2)
	}

	if len(other) > 0 {
		attrs.AddAttribute(models.AttrHeaders, other)
	}
}

func (c *collector) convertActivityID(id string) {
	if c.entities == nil {
		return
	}

	// An activity is an entity with type 'rva'.
	entity := c.entities.GetEntity(fmt.Sprintf("rva::%s", id))
	if entity == nil {
		return
	}

	// It should give us the actual 'doelbinding' from its attributes.
	doel, ok := entity.Attributes().GetAttributeValue(models.AttrDoelbinding).(string)
	if !ok || doel == "" {
		return
	}

	c.parc.Context.AddAttribute(models.AttrDoelbinding, doel)
}
