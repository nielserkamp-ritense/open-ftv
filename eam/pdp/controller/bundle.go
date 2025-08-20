package controller

import (
	"bytes"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// NewBundle implements the Controller interface.
func (b *Base) NewBundle(bundle *bundles.Bundle) (uint64, error) {
	// prevent authorization requests during the bundle processing.
	b.AuthMutex.Lock()
	defer b.AuthMutex.Unlock()

	if err := b.processPolicies(bundle); err != nil {
		return 0, err
	}

	b.processAttributes(bundle)
	b.processEntities(bundle)
	b.processRelations(bundle)

	oldVersion := b.BundleVersion
	b.BundleVersion = bundle.Version
	return oldVersion, nil
}

func (b *Base) processPolicies(bundle *bundles.Bundle) error {
	list := make([]*models.Policy, 0, len(bundle.Policies))
	for _, p1 := range bundle.Policies {
		p2, err := models.NewPolicyFromOAS(p1, bytes.NewBufferString(p1.Data))
		if err != nil {
			return err
		}
		list = append(list, p2)
	}

	return b.PAP.ReplaceAll(list, "*BUNDLE*")
}

func (b *Base) processAttributes(bundle *bundles.Bundle) {
	list := models.NewAttributeSet()
	for _, a := range bundle.Attributes {
		list.AddAttributeKVWithType(a.Key, a.Value, a.Type)
	}

	b.PIP.ReplaceAllAttributes(list, "*BUNDLE*")
}

func (b *Base) processEntities(bundle *bundles.Bundle) {
	list := models.NewEntitySet()
	for _, e := range bundle.Entities {
		attr := models.NewAttributeSet()
		for _, a := range e.Attributes {
			attr.AddAttributeKVWithType(a.Key, a.Value, a.Type)
		}

		list.AddEntity(models.NewEntity(e.Type, e.Id, attr))
	}

	b.PIP.ReplaceAllEntities(list, "*BUNDLE*")
}

func (b *Base) processRelations(_ *bundles.Bundle) {

	// TODO: process relations

}
