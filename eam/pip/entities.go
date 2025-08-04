package pip

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"

// AddEntity adds or replaces an entity in the PIP.
func (p *PIP) AddEntity(entity *models.Entity) {
	prev, ix, err := p.entityPersist.Read(entity.UID())
	if err != nil || prev == nil {
		if _, err = p.entityPersist.Create(entity); err == nil && p.eventSinks != nil {
			p.sendEvent(models.EntityAdded, entity.UID())
		}
	} else {
		if _, err = p.entityPersist.Update(prev, ix, entity); err == nil && p.eventSinks != nil {
			p.sendEvent(models.EntityReplaced, entity.UID())
		}
	}
}

// GetEntity returns an entity from the PIP if it exists.
func (p *PIP) GetEntity(uid string) *models.Entity {
	e, _, _ := p.entityPersist.Read(uid)
	return e
}

// RemoveEntity removes an entity from the PIP.
func (p *PIP) RemoveEntity(uid string) {
	if prev, ix, err := p.entityPersist.Read(uid); err == nil {
		if _, err = p.entityPersist.Delete(prev, ix); err == nil && p.eventSinks != nil {
			p.sendEvent(models.EntityRemoved, uid)
		}
	}
}

// ReplaceAllEntities replaces all entities with the new list.
//
// If an empty list is given, this function effective clears all entities from the PIP.
func (p *PIP) ReplaceAllEntities(list *models.EntitySet) {
	p.IterateEntities(func(a *models.Entity) {
		p.RemoveEntity(a.UID())
	})

	if list != nil {
		p.MergeEntities(list)
	}
}

// IterateEntities calls the given closure for all entities in the PIP.
func (p *PIP) IterateEntities(f models.EntityIterator) {
	if list, err := p.entityPersist.List(); err == nil {
		for i := range list {
			f(list[i])
		}
	}
}

// MergeEntities adds all given entities to the PIP.
func (p *PIP) MergeEntities(in ...*models.EntitySet) {
	for _, set := range in {
		set.IterateEntities(func(e *models.Entity) {
			p.AddEntity(e)
		})
	}
}
