package pip

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"

// AddAttribute implements the AttributeSet interface.
//
// Use this to add a default attribute to the PIP.
func (p *PIP) AddAttribute(key string, value any) {
	_ = p.addAttribute(models.NewAttribute(key, value))
}

// AddAttributeWithType implements the AttributeSet interface.
//
// Use this to add a default attribute to the PIP.
func (p *PIP) AddAttributeWithType(key string, value any, tp string) {
	_ = p.addAttribute(models.NewAttributeWithType(key, value, tp))
}

// AddOriginalAttribute implements the AttributeSet interface.
//
// Use this to add a default attribute to the PIP.
func (p *PIP) AddOriginalAttribute(key string, value, original any, tp string) {
	_ = p.addAttribute(models.NewOriginalAttribute(key, value, original, tp))
}

func (p *PIP) addAttribute(a *models.Attribute) error {
	prev, ix, err := p.attributePersist.Read(a.Key())
	if err != nil || prev == nil {
		if _, err = p.attributePersist.Create(a); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeAdded, a.Key())
		}
	} else {
		if _, err = p.attributePersist.Update(prev, ix, a); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeReplaced, a.Key())
		}
	}
	return err
}

// GetAttribute implements the AttributeSet interface.
//
// Use this to read a default attribute from the PIP.
func (p *PIP) GetAttribute(key string) *models.Attribute {
	a, _, _ := p.attributePersist.Read(key)
	return a
}

// GetAttributeValue implements the AttributeSet interface.
//
// Use this to read a default attribute value from the PIP.
func (p *PIP) GetAttributeValue(key string) any {
	if a, _, _ := p.attributePersist.Read(key); a != nil {
		return a.Value()
	}
	return nil
}

// RemoveAttribute implements the AttributeSet interface.
//
// Use this to remove a default attribute from the PIP.
func (p *PIP) RemoveAttribute(key string) {
	if prev, ix, err := p.attributePersist.Read(key); err == nil {
		if _, err = p.attributePersist.Delete(prev, ix); err == nil && p.eventSinks != nil {
			p.sendEvent(models.AttributeRemoved, key)
		}
	}
}

// ReplaceAllAttributes replaces all attributes with the new list.
//
// If an empty list is given, this function effective clears all attributes from the PIP.
func (p *PIP) ReplaceAllAttributes(list *models.AttributeSet) {
	p.IterateAttributes(func(a *models.Attribute) {
		p.RemoveAttribute(a.Key())
	})

	if list != nil {
		p.MergeAttributes(list)
	}
}

// IterateAttributes implements the AttributeSet interface.
//
// Use this to iterate through all default attributes from the PIP.
func (p *PIP) IterateAttributes(f models.AttributeIterator) {
	if list, err := p.attributePersist.List(); err == nil {
		for i := range list {
			f(list[i])
		}
	}
}

// MergeAttributes implements the AttributeSet interface.
//
// Use this to merge an attribute set into the default attributes of the PIP.
func (p *PIP) MergeAttributes(in ...*models.AttributeSet) {
	for _, set := range in {
		set.IterateAttributes(func(attr *models.Attribute) {
			_ = p.addAttribute(attr)
		})
	}
}
