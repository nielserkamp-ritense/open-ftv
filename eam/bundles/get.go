package bundles

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Get attempts to retrieve the latest deployment for the given bundle id.
func (m *Manager) Get(last *Deployment, id string) (*Bundle, error) {
	cfg, ok := m.bundles[id]
	if !ok {
		return nil, fmt.Errorf("bundle '%s' not found", id)
	}

	out := NewBundle(last.version, cfg.Language, cfg.Tags...)

	if cfg.Policies && m.policies != nil {
		m.policies.Iterate(func(p *models.Policy) {
			out.AddPolicy(p)
		})
	}

	if cfg.Data {
		if m.data != nil {
			m.data.IterateAttributes(func(attr *models.Attribute) {
				out.AddAttribute(attr)
			})

			m.data.IterateEntities(func(e *models.Entity) {
				out.AddEntity(e)
			})

			// m.data.IterateRelations(func(rel *models.Relation) {
			// 	out.AddRelation(rel)
			// })
		}
	}

	return out, nil
}
