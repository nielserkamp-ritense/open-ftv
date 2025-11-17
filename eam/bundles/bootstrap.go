package bundles

import (
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// bootstrapDeployment is used to initiate the very first bundle deployment.
// Since PDPs require a bundle to work, a bootstrap-bundle ensures new installations will have a working setup.
func (m *Manager) bootstrapDeployment() {
	b := &bootstrapper{
		logger:   m.logger,
		policies: m.policies,
		data:     m.data,
		bundles:  m.bundles,
	}
	b.run()
}

type bootstrapper struct {
	m        *Manager
	logger   *slog.Logger
	policies PolicyHandler
	data     DataHandler
	bundles  map[string]*Config
	combos   map[string]fakeParams
	tags     map[string]struct{}
	openfga  bool
}

func (b *bootstrapper) run() {
	b.logger.Info("Initializing bootstrap policies")

	b.uniqueTags()

	for _, params := range b.combos {
		b.createFakePolicy(params)
	}

	if b.openfga {
		b.createRelations()
	}
}

func (b *bootstrapper) uniqueTags() {
	b.combos = make(map[string]fakeParams, len(b.bundles))

	for _, cfg := range b.bundles {
		if cfg.Policies && len(cfg.Tags) > 0 {
			if l := models.LanguageFromString(cfg.Language); l != 0 {
				tag := cfg.Tags[0]

				b.combos[fmt.Sprintf("%d:%s", l, tag)] = fakeParams{language: l, tag: tag}

				if l == models.OPENFGA {
					b.tags[tag] = struct{}{}
					b.openfga = true
				}
			}
		}
	}
}

func (b *bootstrapper) createFakePolicy(params fakeParams) {
	var policy *models.Policy

	switch params.language {
	case models.CEDAR:
		policy = newCedarPolicy(params.tag)
	case models.REGO:
		policy = newRegoPolicy(params.tag)
	case models.CERBOS:
		policy = newCerbosPolicy(params.tag)
	case models.OPENFGA:
		policy = newOpenFGAPolicy(params.tag)
	default:
		return
	}

	policy.WithStatus(models.StatusDeployed)

	if _, err := b.policies.Create(policy, "*BOOTSTRAP*"); err != nil {
		b.logger.Warn("failed to create bootstrap policy", "language", params.language.String(), "tag", params.tag)
	} else {
		b.logger.Info("bootstrap policy created", "language", params.language.String(), "tag", params.tag)
	}
}

func (b *bootstrapper) createRelations() {
	// tags := slices.Sorted(maps.Keys(b.tags))
	// relations := newOpenFGARelations(tags)
	//
	// for _, rel := range relations {
	// 	if _, err := b.data.AddRelation(rel); err != nil {
	// 		b.logger.Warn("failed to add OpenFGA relation", "relation", rel)
	// 	} else {
	// 		b.logger.Info("Bootstrap OpenFGA relations created")
	// 	}
	// }
}

type fakeParams struct {
	language models.Language
	tag      string
}
