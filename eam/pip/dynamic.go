package pip

import (
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// dynamicData contains all dynamically added attributes, entities and relations.
type dynamicData struct {
	modified    bool
	attributes  *models.AttributeSet
	entities    *models.EntitySet
	relations   *models.RelationSet
	dynReporter ReportDynamicData
	mutex       sync.RWMutex
}

func (d *dynamicData) isModified() bool {
	if d.dynReporter == nil {
		return false
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.modified
}

func (d *dynamicData) addAttribute(a *models.Attribute) {
	if d.dynReporter == nil {
		return
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	prev := d.attributes.GetAttribute(a.Key())
	if prev == nil || !prev.Equals(a) {
		_, _ = d.attributes.AddAttribute(a)
		d.modified = true
	}
}

func (d *dynamicData) addEntity(e *models.Entity) {
	if d.dynReporter == nil {
		return
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	prev := d.entities.GetEntity(e.UID())
	if prev == nil || !prev.Equals(e) {
		_, _ = d.entities.AddEntity(e)
		d.modified = true
	}
}

func (d *dynamicData) addRelation(r *models.Relation) {
	if d.dynReporter == nil {
		return
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	prev := d.relations.GetRelation(r.UID())
	if prev == nil || !prev.Equals(r) {
		_, _ = d.relations.AddRelation(r)
		d.modified = true
	}
}

func (d *dynamicData) report(f func()) {
	if d.dynReporter == nil {
		f()
		return
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.modified {
		f()
		return
	}

	m1 := models.MapFromAttributes(d.attributes)

	m2 := make(map[string]*models.Entity, 16)
	d.entities.IterateEntities(func(e *models.Entity) {
		m2[e.UID()] = e
	})

	m3 := make(map[string]*models.Relation, 16)
	d.relations.IterateRelations(func(r *models.Relation) {
		m3[r.UID()] = r
	})

	m := make(map[string]any, 4)
	if len(m1) > 0 {
		m["attributes"] = m1
	}
	if len(m2) > 0 {
		m["entities"] = m2
	}
	if len(m3) > 0 {
		m["relations"] = m3
	}

	d.dynReporter(m)
	d.modified = false

	f()
}
