package controller

import (
	"bytes"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

var bundleUser = identity.NewBundlePrincipal()

// NewBundle implements the Controller interface.
func (b *Base) NewBundle(bundle *bundles.Bundle) (uint64, error) {
	// prevent authorization requests during the bundle processing.
	b.AuthMutex.Lock()
	defer b.AuthMutex.Unlock()

	if b.ADL != nil {
		pip.WithDynamicReporter(b.ADL.NewInformation)(b.PIP)
	}

	if err := b.processPolicies(bundle); err != nil {
		return 0, err
	}

	b.processAttributes(bundle)
	b.processEntities(bundle)
	b.processRelations(bundle)

	oldVersion := b.BundleVersion
	b.BundleVersion = bundle.Version

	if b.ADL != nil {
		pip.WithDynamicReporter(b.ADL.NewInformation)(b.PIP)
		b.ADL.NewBundle(b.BundleVersion)
	}

	return oldVersion, nil
}

func (b *Base) processPolicies(bundle *bundles.Bundle) error {
	list := make([]*models.Policy, 0, len(bundle.Policies))
	for _, p1 := range bundle.Policies {
		if p1.Language == "" {
			p1.Language = bundle.Language
		}
		p2, err := models.NewPolicyFromOAS(p1, bytes.NewBufferString(p1.Data))
		if err != nil {
			return err
		}
		list = append(list, p2)
	}

	return b.PAP.ReplaceAll(list, bundleUser)
}

// processAttributes upserts the bundle's attributes into the PIP, and removes only the
// attributes the previous bundle added that are no longer present in this one. Attributes owned
// by other PIP sources (a local file store, a pull-config) are never touched.
func (b *Base) processAttributes(bundle *bundles.Bundle) {
	newKeys := make(map[string]struct{}, len(bundle.Attributes))

	for _, a := range bundle.Attributes {
		attr := models.NewAttributeWithType(a.Key, a.Value, a.Type)
		_, _ = b.PIP.AddAttributeWithUser(attr, bundleUser)

		newKeys[a.Key] = struct{}{}
	}

	for key := range b.bundleAttrKeys {
		if _, keep := newKeys[key]; !keep {
			_, _ = b.PIP.RemoveAttribute(key, bundleUser)
		}
	}

	b.bundleAttrKeys = newKeys
}

// processEntities upserts the bundle's entities into the PIP, and removes only the entities the
// previous bundle added that are no longer present in this one. Entities owned by other PIP
// sources (a local file store, a pull-config) are never touched.
func (b *Base) processEntities(bundle *bundles.Bundle) {
	newUIDs := make(map[string]struct{}, len(bundle.Entities))

	for _, e := range bundle.Entities {
		attr := models.NewAttributeSet()
		for _, a := range e.Attributes {
			attr.AddAttributeKVWithType(a.Key, a.Value, a.Type)
		}

		entity := models.NewEntity(e.Type, e.Id, attr)
		_, _ = b.PIP.AddEntityWithUser(entity, bundleUser)

		newUIDs[entity.UID()] = struct{}{}
	}

	for uid := range b.bundleEntityUIDs {
		if _, keep := newUIDs[uid]; !keep {
			_, _ = b.PIP.RemoveEntity(uid, bundleUser)
		}
	}

	b.bundleEntityUIDs = newUIDs
}

func (b *Base) processRelations(_ *bundles.Bundle) {

	// TODO: process relations

}
