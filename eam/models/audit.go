package models

import "time"

// Audit contains the audit details for an object.
//
// It is meant to be embedded in object structs that require audit details.
type Audit struct {
	created   time.Time
	createdBy string
	updated   time.Time
	updatedBy string
}

// Created returns the creation timestamp of the object.
func (a *Audit) Created() time.Time {
	return a.created
}

// CreatedBy returns the user who created the object.
func (a *Audit) CreatedBy() string {
	return a.createdBy
}

// Updated returns the last update timestamp of the object.
func (a *Audit) Updated() time.Time {
	return a.updated
}

// UpdatedBy returns the user who last updated the object.
func (a *Audit) UpdatedBy() string {
	return a.updatedBy
}
