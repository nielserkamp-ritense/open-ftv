package pip

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// AddRelation adds/replaces the given attribute in the PIP.
//
// If the given attribute exists, it is updated, otherwise created.
func (p *PIP) AddRelation(in *models.Relation) (*models.Relation, error) {
	// return p.addRelationWithUser(in, "")
	return in, nil
}

// AddDynamicRelation adds/replaces the given attribute in the PIP.
//
// Unlike AddRelation, this function will also mark the attribute as dynamic with respect to the Authorization Decision Log.
//
// If the given attribute exists, it is updated, otherwise created.
// func (p *PIP) AddDynamicRelation(in *models.Relation) (*models.Relation, error) {
// 	p.dynamicData.addRelation(in)
// 	return p.addRelationWithUser(in, "")
// }

// AddRelationFromOAS adds/replaces an attribute in the PIP based on the given OAS model.
//
// Unlike AddRelation, this function will also mark the attribute as dynamic with respect to the Authorization Decision Log.
//
// If the given attribute exists, it is updated, otherwise created.
// func (p *PIP) AddRelationFromOAS(in *oas.Relation, user string) (*models.Relation, error) {
// 	a := models.NewRelationFromOAS(in)
// 	p.dynamicData.addRelation(a)
// 	return p.addRelationWithUser(a, user)
// }

// func (p *PIP) addRelationWithUser(in *models.Relation, user string) (*models.Relation, error) {
// 	prev, ix, err := p.attributeDB.ReadRelation(p.ctx, in.Key())
// 	if err != nil || prev == nil {
// 		if _, err = p.attributeDB.CreateRelation(context.WithValue(p.ctx, "user", user), in); err == nil && p.eventSinks != nil {
// 			p.sendEvent(models.RelationAdded, in.Key())
// 		}
// 	} else {
// 		if _, err = p.attributeDB.UpdateRelation(context.WithValue(p.ctx, "user", user), prev, ix, in); err == nil && p.eventSinks != nil {
// 			p.sendEvent(models.RelationReplaced, in.Key())
// 		}
// 	}
// 	return in, err
// }

// AddRelationKV adds/replaces an attribute in the PIP with the specified key and value.
// func (p *PIP) AddRelationKV(key string, value any) (*models.Relation, error) {
// 	return p.AddRelation(models.NewRelation(key, value))
// }

// AddRelationKVWithType adds/replaces an attribute in the PIP with a specified key, value and type.
// func (p *PIP) AddRelationKVWithType(key string, value any, tp string) (*models.Relation, error) {
// 	return p.AddRelation(models.NewRelationWithType(key, value, tp))
// }

// AddOriginalRelation adds/replaces an attribute in the PIP with a specified key, value, type and original value.
// func (p *PIP) AddOriginalRelation(key string, value, original any, tp string) (*models.Relation, error) {
// 	return p.AddRelation(models.NewOriginalRelation(key, value, original, tp))
// }

// GetRelation retrieves an attribute from the PIP.
// func (p *PIP) GetRelation(key string) (*models.Relation, uint64, error) {
// 	return p.attributeDB.ReadRelation(p.ctx, key)
// }

// GetRelationAudit retrieves the audit-log of an attribute from the PIP.
// func (p *PIP) GetRelationAudit(key string) ([]oas.AuditEntry, error) {
// 	return p.attributeDB.ReadRelationAudit(p.ctx, key)
// }

// GetRelationDeployments retrieves the deployment-log of an attribute from the PIP.
// func (p *PIP) GetRelationDeployments(key string) ([]oas.UsageData, error) {
// 	return p.attributeDB.ReadRelationDeployments(p.ctx, key)
// }

// GetRelationValue retrieves the value of an attribute value from the PIP.
// func (p *PIP) GetRelationValue(key string) any {
// 	if a, _, _ := p.attributeDB.ReadRelation(p.ctx, key); a != nil {
// 		return a.Value()
// 	}
// 	return nil
// }

// RemoveRelation removes an attribute from the PIP.
// func (p *PIP) RemoveRelation(key, user string) (*models.Relation, error) {
// 	p.dynamicData.attributes.RemoveRelation(key) // also remove from dynamic data.
// 	return p.removeRelation(key, user)
// }

// func (p *PIP) removeRelation(key, user string) (*models.Relation, error) {
// 	prev, ix, err := p.attributeDB.ReadRelation(p.ctx, key)
// 	if err == nil {
// 		if _, err = p.attributeDB.DeleteRelation(context.WithValue(p.ctx, "user", user), prev, ix); err == nil && p.eventSinks != nil {
// 			p.sendEvent(models.RelationRemoved, key)
// 		}
// 	}
// 	return prev, err
// }

// ReplaceAllRelations replaces all attributes with the new list.
//
// If an empty list is given, this function effectively clears all attributes from the PIP.
// func (p *PIP) ReplaceAllRelations(list *models.RelationSet, user string) {
// 	// delete all existing attributes.
// 	keys := make([]string, 0)
// 	p.IterateRelations(func(a *models.Relation) {
// 		keys = append(keys, a.Key())
// 	})
//
// 	for i := range keys {
// 		_, _ = p.removeRelation(keys[i], user)
// 	}
//
// 	if list != nil {
// 		// add all given attributes.
// 		list.IterateRelations(func(a *models.Relation) {
// 			_, _ = p.addRelationWithUser(a, user)
// 		})
// 	}
//
// 	p.MergeRelations(p.dynamicData.attributes)
// }

// IterateRelations calls the given closure for all attributes in the PIP.
func (p *PIP) IterateRelations(f models.RelationIterator) {
	// if list, err := p.attributeDB.ListRelations(p.ctx); err == nil {
	// 	for i := range list {
	// 		f(list[i])
	// 	}
	// }
}

// MergeRelations merges the given attribute set(s) into the PIP.
// func (p *PIP) MergeRelations(in ...*models.RelationSet) {
// 	for _, set := range in {
// 		set.IterateRelations(func(attr *models.Relation) {
// 			_, _ = p.addRelationWithUser(attr, "")
// 		})
// 	}
// }
