package pip

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// AddAttribute adds/replaces the given attribute in the PIP.
//
// If the given attribute exists, it is updated, otherwise created.
func (p *PIP) AddAttribute(in *models.Attribute) (*models.Attribute, error) {
	return p.addAttributeWithUser(in, "")
}

// AddAttributeFromOAS adds/replaces an attribute in the PIP based on the given OAS model.
func (p *PIP) AddAttributeFromOAS(in *oas.Attribute, user string) (*models.Attribute, error) {
	return p.addAttributeWithUser(models.NewAttributeFromOAS(in), user)
}

func (p *PIP) addAttributeWithUser(in *models.Attribute, user string) (*models.Attribute, error) {
	prev, ix, err := p.attributeDB.ReadAttribute(p.ctx, in.Key())
	if err != nil || prev == nil {
		if _, err = p.attributeDB.CreateAttribute(context.WithValue(p.ctx, "user", user), in); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeAdded, in.Key())
		}
	} else {
		if _, err = p.attributeDB.UpdateAttribute(context.WithValue(p.ctx, "user", user), prev, ix, in); err == nil && p.eventSinks != nil {
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

// GetAttributeValue retrieves the value of an attribute value from the PIP.
func (p *PIP) GetAttributeValue(key string) any {
	if a, _, _ := p.attributeDB.ReadAttribute(p.ctx, key); a != nil {
		return a.Value()
	}
	return nil
}

// RemoveAttribute removes an attribute from the PIP.
func (p *PIP) RemoveAttribute(key, user string) (*models.Attribute, error) {
	prev, ix, err := p.attributeDB.ReadAttribute(p.ctx, key)
	if err == nil {
		if _, err = p.attributeDB.DeleteAttribute(context.WithValue(p.ctx, "user", user), prev, ix); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeRemoved, key)
		}
	}
	return prev, err
}

// ReplaceAllAttributes replaces all attributes with the new list.
//
// If an empty list is given, this function effectively clears all attributes from the PIP.
func (p *PIP) ReplaceAllAttributes(list *models.AttributeSet, user string) {
	// delete all existing attributes.
	keys := make([]string, 0)
	p.IterateAttributes(func(a *models.Attribute) {
		keys = append(keys, a.Key())
	})

	for i := range keys {
		_, _ = p.RemoveAttribute(keys[i], user)
	}

	if list != nil {
		// add all given attributes.
		list.IterateAttributes(func(a *models.Attribute) {
			_, _ = p.addAttributeWithUser(a, user)
		})
	}
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
			_, _ = p.AddAttribute(attr)
		})
	}
}
