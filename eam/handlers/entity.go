package handlers

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/attributes"
)

// EntityFromOAS converts an OAS entity model to an internal model.
func EntityFromOAS(in *attributes.Entity, attrs models.AttributeSet) models.Entity {
	for i := range in.Attributes {
		a := AttributeFromOAS(&in.Attributes[i])
		attrs.AddAttributeWithType(a.Key(), a.Value(), a.Type())
	}
	return models.NewEntity(in.Type, in.Id, attrs)
}

// EntityToOAS converts an internal entity model to an OAS model.
func EntityToOAS(in models.Entity) *attributes.Entity {
	out := &attributes.Entity{
		Type:       in.Type(),
		Id:         in.ID(),
		Attributes: make([]attributes.Attribute, 0),
	}

	in.Attributes().IterateAttributes(func(attr models.Attribute) {
		a := AttributeToOAS(attr)
		out.Attributes = append(out.Attributes, *a)
	})

	return out
}
