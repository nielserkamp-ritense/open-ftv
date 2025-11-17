package pip

import (
	"context"
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// AddEntity adds/replaces an entity in the PIP.
//
// If the given attribute exists, it is updated, otherwise created.
func (p *PIP) AddEntity(entity *models.Entity) (*models.Entity, error) {
	return p.AddEntityWithUser(entity, "")
}

// AddDynamicEntity adds/replaces an entity in the PIP.
//
// Unlike AddEntity, this function will also mark the entity as dynamic with respect to the Authorization Decision Log.
//
// If the given attribute exists, it is updated, otherwise created.
func (p *PIP) AddDynamicEntity(entity *models.Entity) (*models.Entity, error) {
	p.dynamicData.addEntity(entity)
	return p.AddEntityWithUser(entity, "")
}

// AddEntityFromOAS adds/replaces an entity in the PIP based on the given OAS model.
//
// Unlike AddEntity, this function will also mark the entity as dynamic with respect to the Authorization Decision Log.
//
// If the given attribute exists, it is updated, otherwise created.
func (p *PIP) AddEntityFromOAS(in *oas.Entity, user string) (*models.Entity, error) {
	e := models.EntityFromOAS(in)
	p.dynamicData.addEntity(e)
	return p.AddEntityWithUser(e, user)
}

// AddEntityWithUser adds/replaces an entity in the PIP on behalf of the given user.
func (p *PIP) AddEntityWithUser(entity *models.Entity, user string) (*models.Entity, error) {
	e2, ix, err := p.entityDB.ReadEntity(p.ctx, entity.Type(), entity.ID())
	if err != nil || e2 == nil {
		if e2, err = p.entityDB.CreateEntity(context.WithValue(p.ctx, "user", user), entity); err == nil && p.eventSinks != nil {
			p.sendEvent(models.EntityAdded, entity.UID())
		}
	} else {
		if e2, err = p.entityDB.UpdateEntity(context.WithValue(p.ctx, "user", user), e2, ix, entity); err == nil && p.eventSinks != nil {
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

// GetEntityAudit retrieves the audit-log of an attribute from the PIP.
func (p *PIP) GetEntityAudit(uid string) ([]oas.AuditEntry, error) {
	ns, id := models.SplitEntityUID(uid)
	return p.entityDB.ReadEntityAudit(p.ctx, ns, id)
}

// GetEntityDeployments retrieves the deployment-log of an attribute from the PIP.
func (p *PIP) GetEntityDeployments(uid string) ([]oas.UsageData, error) {
	ns, id := models.SplitEntityUID(uid)
	return p.entityDB.ReadEntityDeployments(p.ctx, ns, id)
}

// UpdateEntityStatus updates the status of an attribute in cache/storage.
//
// An error is returned if the attribute key doesn't exist or the status update is not allowed.
func (p *PIP) UpdateEntityStatus(uid string, status models.Status, user string) (*models.Entity, error) {
	ns, id := models.SplitEntityUID(uid)
	prev, _, err := p.entityDB.ReadEntity(p.ctx, ns, id)
	if err != nil || prev == nil {
		return nil, fmt.Errorf("entity not found")
	}

	if (prev.Status() == models.StatusConcept && status == models.StatusAccepted) ||
		(prev.Status() == models.StatusAccepted && status == models.StatusConcept) {
		return p.AddEntity(prev.WithStatus(status))
	}
	return nil, fmt.Errorf("invalid status change from %s to %s", prev.Status().String(), status.String())
}

// RemoveEntity removes an entity from the PIP.
func (p *PIP) RemoveEntity(uid string, user string) (*models.Entity, error) {
	p.dynamicData.entities.RemoveEntity(uid)
	return p.removeEntity(uid, user)
}

func (p *PIP) removeEntity(uid string, user string) (*models.Entity, error) {
	ns, id := models.SplitEntityUID(uid)
	e2, ix, err := p.entityDB.ReadEntity(p.ctx, ns, id)
	if err == nil {
		if e2, err = p.entityDB.DeleteEntity(context.WithValue(p.ctx, "user", user), e2, ix); err == nil && p.eventSinks != nil {
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
		_, _ = p.removeEntity(a.UID(), user)
	})

	// add all given entities.
	if list != nil {
		list.IterateEntities(func(e *models.Entity) {
			_, _ = p.AddEntityWithUser(e, user)
		})
	}

	p.MergeEntities(p.dynamicData.entities)
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
