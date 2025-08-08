package pip

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// AddAttribute adds the given attribute to the PIP.
func (p *PIP) AddAttribute(in *models.Attribute) (*models.Attribute, error) {
	prev, ix, err := p.attributePersist.Read(in.Key())
	if err != nil || prev == nil {
		if _, err = p.attributePersist.Create(in); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeAdded, in.Key())
		}
	} else {
		if _, err = p.attributePersist.Update(prev, ix, in); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeReplaced, in.Key())
		}
	}
	return in, err
}

// AddAttributeFromOAS adds an attribute to the PIP based on the given OAS model.
func (p *PIP) AddAttributeFromOAS(in *attributes.Attribute) (*models.Attribute, error) {
	return p.AddAttribute(models.NewAttributeFromOAS(in))
}

// AddAttributeKV adds an attribute to the PIP with the specified key and value.
func (p *PIP) AddAttributeKV(key string, value any) (*models.Attribute, error) {
	return p.AddAttribute(models.NewAttribute(key, value))
}

// AddAttributeKVWithType adds an attribute to the PIP with a specified key, value and type.
func (p *PIP) AddAttributeKVWithType(key string, value any, tp string) (*models.Attribute, error) {
	return p.AddAttribute(models.NewAttributeWithType(key, value, tp))
}

// AddOriginalAttribute adds an attribute to the PIP with a specified key, value, type and original value.
func (p *PIP) AddOriginalAttribute(key string, value, original any, tp string) (*models.Attribute, error) {
	return p.AddAttribute(models.NewOriginalAttribute(key, value, original, tp))
}

// GetAttribute retrieves an attribute from the PIP.
func (p *PIP) GetAttribute(key string) *models.Attribute {
	a, _, _ := p.attributePersist.Read(key)
	return a
}

// GetAttributeValue retrieves the value of an attribute value from the PIP.
func (p *PIP) GetAttributeValue(key string) any {
	if a, _, _ := p.attributePersist.Read(key); a != nil {
		return a.Value()
	}
	return nil
}

// RemoveAttribute removes an attribute from the PIP.
func (p *PIP) RemoveAttribute(key string) (*models.Attribute, error) {
	prev, ix, err := p.attributePersist.Read(key)
	if err == nil {
		if _, err = p.attributePersist.Delete(prev, ix); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeRemoved, key)
		}
	}
	return prev, err
}

// ReplaceAllAttributes replaces all attributes with the new list.
//
// If an empty list is given, this function effectively clears all attributes from the PIP.
func (p *PIP) ReplaceAllAttributes(list *models.AttributeSet) {
	keys := make([]string, 0)

	p.IterateAttributes(func(a *models.Attribute) {
		keys = append(keys, a.Key())
	})

	for i := range keys {
		_, _ = p.RemoveAttribute(keys[i])
	}

	if list != nil {
		p.MergeAttributes(list)
	}
}

// IterateAttributes calls the given closure for all attributes in the PIP.
func (p *PIP) IterateAttributes(f models.AttributeIterator) {
	if list, err := p.attributePersist.List(); err == nil {
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
