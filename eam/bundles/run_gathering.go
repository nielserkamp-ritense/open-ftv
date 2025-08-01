package bundles

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func (r *runner) gathering() {
	r.debug("gather stage started")

	r.gatherLists("gather")

	if r.advance("gather") {
		r.info("gather stage statistics", "policies", len(r.policies), "attributes", len(r.attributes), "entities", len(r.entities), "relations", len(r.relations))
	}
}

func (r *runner) gatherLists(stage string) {
	r.gatherPolicies(stage)
	r.gatherAttributes()
	r.gatherEntities()
	r.gatherRelations()
}

func (r *runner) gatherPolicies(stage string) {
	if r.bundles == nil {
		r.initCountsAndSlices()
		r.info(fmt.Sprintf("%s stage restart", stage), "bundles", r.bundleCount, "targets", r.targetCount)
	}

	if r.m.policies == nil {
		return
	}

	r.m.policies.Iterate(func(p *models.Policy) {
		r.policies = append(r.policies, p)
	})
}

func (r *runner) gatherAttributes() {
	if r.m.attributes == nil {
		return
	}

	r.m.attributes.IterateAttributes(func(attr *models.Attribute) {
		r.attributes = append(r.attributes, attr)
	})
}

func (r *runner) gatherEntities() {
	if r.m.entities == nil {
		return
	}

	r.m.entities.IterateEntities(func(e *models.Entity) {
		r.entities = append(r.entities, e)
	})
}

func (r *runner) gatherRelations() {
	if r.m.relations == nil {
		return
	}

	r.m.relations.IterateRelations(func(rel *models.Relation) {
		r.relations = append(r.relations, rel)
	})
}
