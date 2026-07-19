package pip

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// AddAttribute adds/replaces the given attribute in the PIP.
//
// If the given attribute exists, it is updated, otherwise created.
func (p *PIP) AddAttribute(in *models.Attribute) (*models.Attribute, error) {
	return p.addAttributeWithUser(in, identity.NewSystemPrincipal())
}

// AddDynamicAttribute adds/replaces the given attribute in the PIP.
//
// Unlike AddAttribute, this function will also mark the attribute as dynamic with respect to the Authorization Decision Log.
//
// If the given attribute exists, it is updated, otherwise created.
func (p *PIP) AddDynamicAttribute(in *models.Attribute) (*models.Attribute, error) {
	p.dynamicData.addAttribute(in)
	return p.addAttributeWithUser(in, identity.NewSystemPrincipal())
}

// AddAttributeFromOAS adds/replaces an attribute in the PIP based on the given OAS model.
//
// Unlike AddAttribute, this function will also mark the attribute as dynamic with respect to the Authorization Decision Log.
//
// If the given attribute exists, it is updated, otherwise created.
func (p *PIP) AddAttributeFromOAS(in *oas.Attribute, user identity.Principal) (*models.Attribute, error) {
	a := models.NewAttributeFromOAS(in)
	p.dynamicData.addAttribute(a)
	return p.addAttributeWithUser(a, user)
}

func (p *PIP) addAttributeWithUser(in *models.Attribute, user identity.Principal) (*models.Attribute, error) {
	prev, ix, err := p.attributeDB.ReadAttribute(p.ctx, in.Key())
	if err != nil || prev == nil {
		if _, err = p.attributeDB.CreateAttribute(p.ctx, user, in); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeAdded, in.Key())
		}
	} else {
		if _, err = p.attributeDB.UpdateAttribute(p.ctx, user, prev, ix, in); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeReplaced, in.Key())
		}
	}
	return in, err
}

// AddAttributeKV adds/replaces an attribute in the PIP with the specified key and value.
func (p *PIP) AddAttributeKV(key string, value any) (*models.Attribute, error) {
	return p.AddAttribute(models.NewAttribute(key, value))
}

// AddAttributeKVWithType adds/replaces an attribute in the PIP with a specified key, value and type.
func (p *PIP) AddAttributeKVWithType(key string, value any, tp string) (*models.Attribute, error) {
	return p.AddAttribute(models.NewAttributeWithType(key, value, tp))
}

// AddOriginalAttribute adds/replaces an attribute in the PIP with a specified key, value, type and original value.
func (p *PIP) AddOriginalAttribute(key string, value, original any, tp string) (*models.Attribute, error) {
	return p.AddAttribute(models.NewOriginalAttribute(key, value, original, tp))
}

// GetAttribute retrieves an attribute from the PIP.
func (p *PIP) GetAttribute(key string) (*models.Attribute, uint64, error) {
	return p.attributeDB.ReadAttribute(p.ctx, key)
}

// GetAttributeAudit retrieves the audit-log of an attribute from the PIP.
func (p *PIP) GetAttributeAudit(key string) ([]oas.AuditEntry, error) {
	return p.attributeDB.ReadAttributeAudit(p.ctx, key)
}

// GetAttributeDeployments retrieves the deployment-log of an attribute from the PIP.
func (p *PIP) GetAttributeDeployments(key string) ([]oas.UsageData, error) {
	return p.attributeDB.ReadAttributeDeployments(p.ctx, key)
}

// GetAttributeVersions retrieves the versions of an attribute from the PIP.
func (p *PIP) GetAttributeVersions(key string) (oas.AttributeVersions, error) {
	return p.attributeDB.ReadAttributeVersions(p.ctx, key)
}

// GetAttributeVersion retrieves a specific version of an attribute from the PIP.
func (p *PIP) GetAttributeVersion(key string, version int) (*oas.AttributeVersion, error) {
	return p.attributeDB.ReadAttributeVersion(p.ctx, key, version)
}

// GetAttributeValue retrieves the value of an attribute value from the PIP.
func (p *PIP) GetAttributeValue(key string) any {
	if a, _, _ := p.attributeDB.ReadAttribute(p.ctx, key); a != nil {
		return a.Value()
	}
	return nil
}

// RestoreAttributeVersion restores a specific version of an attribute as the current concept.
func (p *PIP) RestoreAttributeVersion(key string, version int, user identity.Principal) (*models.Attribute, error) {
	attrOld, err := p.attributeDB.ReadAttributeVersion(p.ctx, key, version)
	if err != nil {
		return nil, err
	}

	attr, lastIndex, err2 := p.attributeDB.ReadAttribute(p.ctx, key)
	if err2 != nil {
		return nil, err2
	}

	attr = attr.RestoreFrom(attrOld)
	return p.UpdateAttribute(attr, lastIndex, attr, user)
}

// UpdateAttribute modifies an attribute in cache/storage with a newer version.
//
// An error is returned if the attribute-key doesn't exist.
func (p *PIP) UpdateAttribute(prev *models.Attribute, lastIndex uint64, in *models.Attribute, user identity.Principal) (out *models.Attribute, err error) {
	if out, err = p.attributeDB.UpdateAttribute(p.ctx, user, prev, lastIndex, in); err == nil && out != nil && p.eventSinks != nil {
		p.sendEvent(models.AttributeReplaced, out.Key())
	}
	return
}

// UpdateAttributeStatus updates the status of an attribute in cache/storage.
//
// An error is returned if the attribute key doesn't exist or the status update is not allowed.
func (p *PIP) UpdateAttributeStatus(key string, status models.Status, user identity.Principal) (*models.Attribute, error) {
	prev, _, err := p.attributeDB.ReadAttribute(p.ctx, key)
	if err != nil || prev == nil {
		return nil, fmt.Errorf("attribute not found")
	}

	if (prev.Status() == models.StatusConcept && status == models.StatusAccepted) ||
		(prev.Status() == models.StatusAccepted && status == models.StatusConcept) {
		return p.addAttributeWithUser(prev.WithStatus(status), user)
	}
	return nil, fmt.Errorf("invalid status change from %s to %s", prev.Status().String(), status.String())
}

// RemoveAttribute removes an attribute from the PIP.
func (p *PIP) RemoveAttribute(key string, user identity.Principal) (*models.Attribute, error) {
	p.dynamicData.attributes.RemoveAttribute(key) // also remove from dynamic data.
	return p.removeAttribute(key, user)
}

func (p *PIP) removeAttribute(key string, user identity.Principal) (*models.Attribute, error) {
	prev, ix, err := p.attributeDB.ReadAttribute(p.ctx, key)
	if err == nil {
		if _, err = p.attributeDB.DeleteAttribute(p.ctx, user, prev, ix); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeRemoved, key)
		}
	}
	return prev, err
}

// ReplaceAllAttributes replaces all attributes with the new list.
//
// If an empty list is given, this function effectively clears all attributes from the PIP.
func (p *PIP) ReplaceAllAttributes(list *models.AttributeSet, user identity.Principal) {
	// delete all existing attributes.
	keys := make([]string, 0)
	p.IterateAttributes(func(a *models.Attribute) {
		keys = append(keys, a.Key())
	})

	for i := range keys {
		_, _ = p.removeAttribute(keys[i], user)
	}

	if list != nil {
		// add all given attributes.
		list.IterateAttributes(func(a *models.Attribute) {
			_, _ = p.addAttributeWithUser(a, user)
		})
	}

	p.MergeAttributes(p.dynamicData.attributes)
}

// IterateAttributes calls the given closure for all attributes in the PIP.
func (p *PIP) IterateAttributes(f models.AttributeIterator) {
	if list, err := p.attributeDB.ListAttributes(p.ctx); err == nil {
		for i := range list {
			f(list[i])
		}
	}
}

// MergeAttributes merges the given attribute set(s) into the PIP.
func (p *PIP) MergeAttributes(in ...*models.AttributeSet) {
	for _, set := range in {
		set.IterateAttributes(func(attr *models.Attribute) {
			_, _ = p.addAttributeWithUser(attr, identity.NewSystemPrincipal())
		})
	}
}
