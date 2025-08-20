package pip

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"

// AddEntity adds/replaces an entity in the PIP.
func (p *PIP) AddEntity(entity *models.Entity) (*models.Entity, error) {
	return p.AddEntityWithUser(entity, "")
}

// AddEntityWithUser adds/replaces an entity in the PIP on behalf of the given user.
func (p *PIP) AddEntityWithUser(entity *models.Entity, user string) (*models.Entity, error) {
	e2, ix, err := p.entityDB.ReadEntity(p.ctx, entity.Type(), entity.ID())
	if err != nil || e2 == nil {
		if e2, err = p.entityDB.CreateEntity(p.ctx, user, entity); err == nil && p.eventSinks != nil {
			p.sendEvent(models.EntityAdded, entity.UID())
		}
	} else {
		if e2, err = p.entityDB.UpdateEntity(p.ctx, user, e2, ix, entity); err == nil && p.eventSinks != nil {
			p.sendEvent(models.EntityReplaced, entity.UID())
		}
	}
	return e2, err
}

// GetEntity returns an entity from the PIP if it exists.
func (p *PIP) GetEntity(uid string) (*models.Entity, uint64, error) {
	ns, id := models.SplitEntityUID(uid)
	return p.entityDB.ReadEntity(p.ctx, ns, id)
}

// RemoveEntity removes an entity from the PIP.
func (p *PIP) RemoveEntity(uid string, user string) (*models.Entity, error) {
	ns, id := models.SplitEntityUID(uid)
	e2, ix, err := p.entityDB.ReadEntity(p.ctx, ns, id)
	if err == nil {
		if e2, err = p.entityDB.DeleteEntity(p.ctx, user, e2, ix); err == nil && p.eventSinks != nil {
			p.sendEvent(models.EntityRemoved, uid)
		}
	}
	return e2, err
}

// ReplaceAllEntities replaces all entities with the new list.
//
// If an empty list is given, this function effective clears all entities from the PIP.
func (p *PIP) ReplaceAllEntities(list *models.EntitySet, user string) {
	// remove all existing entities.
	p.IterateEntities(func(a *models.Entity) {
		_, _ = p.RemoveEntity(a.UID(), user)
	})

	// add all given entities.
	if list != nil {
		list.IterateEntities(func(e *models.Entity) {
			_, _ = p.AddEntityWithUser(e, user)
		})
	}
}

// IterateEntities calls the given closure for all entities in the PIP.
func (p *PIP) IterateEntities(f models.EntityIterator) {
	if list, err := p.entityDB.ListEntities(p.ctx); err == nil {
		for i := range list {
			f(list[i])
		}
	}
}

// MergeEntities adds all given entity sets to the PIP.
func (p *PIP) MergeEntities(in ...*models.EntitySet) {
	for _, set := range in {
		set.IterateEntities(func(e *models.Entity) {
			_, _ = p.AddEntity(e)
		})
	}
}
