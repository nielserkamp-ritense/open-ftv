package pep

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
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
				attrs.AddAttributeKV(models.AttrContentType, first)
			case models.HeaderAuthorization:
				c.processAuth(first)
			case models.HeaderFSCAuthorization:
				c.processFSC(first)
			case models.HeaderAPIKey, "apikey", "x-apikey", "x-api-key":
				attrs.AddAttributeKV(models.AttrAPIKey, first)
			case models.HeaderRvvaID, models.HeaderObsoleteRvvaID:
				activityID = first
				attrs.AddAttributeKV(models.AttrRvvaID, activityID)
				c.convertActivityID(activityID)
			case models.HeaderCoreUser, models.HeaderObsoleteCoreUser:
				attrs.AddAttributeKV(models.AttrCoreUser, first)
			case models.HeaderGrondslag:
				attrs.AddAttributeKV(models.AttrGrondslag, first)
			case models.HeaderDoelbinding:
				attrs.AddAttributeKV(models.AttrDoelbinding, first)
			case models.HeaderZaakType, "zaaktype":
				attrs.AddAttributeKV(models.AttrZaakType, first)
			case models.HeaderTaak:
				attrs.AddAttributeKV(models.AttrTaak, first)
			case models.HeaderXForwardedFor:
				fwd1 = strings.Join(list, ",")
			case models.HeaderForwarded:
				fwd2 = strings.Join(list, ",")
			case models.HeaderDeviceID, models.HeaderObsoleteDeviceID:
				c.parc.Principal.Attributes().AddAttributeKV(models.AttrDeviceID, first)
			case models.HeaderTraceParent:
				attrs.AddAttributeKV(models.AttrTraceParent, first)
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
		attrs.AddAttributeKV(models.AttrHeaders, other)
	}
}

func (c *collector) convertActivityID(id string) {
	if c.getEntity == nil {
		return
	}

	// An activity is an entity with type 'rva'.
	entity, _, _ := c.getEntity(fmt.Sprintf("rva::%s", id))
	if entity == nil {
		return
	}

	// It should give us the actual 'doelbinding' from its attributes.
	doel, ok := entity.Attributes().GetAttributeValue(models.AttrDoelbinding).(string)
	if !ok || doel == "" {
		return
	}

	c.parc.Context.AddAttributeKV(models.AttrDoelbinding, doel)
}
