package network

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// recordingAttributes wraps an AttributeSet and reports every mutated key to a callback.
type recordingAttributes struct {
	models.AttributeSet
	rec func(key string)
}

func (r *recordingAttributes) AddAttribute(key string, value any) {
	r.AttributeSet.AddAttribute(key, value)
	r.rec(key)
}

func (r *recordingAttributes) AddAttributeWithType(key string, value any, tp string) {
	r.AttributeSet.AddAttributeWithType(key, value, tp)
	r.rec(key)
}

func (r *recordingAttributes) AddOriginalAttribute(key string, value, original any, tp string) {
	r.AttributeSet.AddOriginalAttribute(key, value, original, tp)
	r.rec(key)
}

// recordingEntities wraps an EntitySet and reports every added entity to a callback.
type recordingEntities struct {
	models.EntitySet
	rec func(entity models.Entity)
}

func (r *recordingEntities) AddEntity(entity models.Entity) {
	r.EntitySet.AddEntity(entity)
	r.rec(entity)
}
