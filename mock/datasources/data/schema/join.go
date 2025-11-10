package schema

import (
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

// Join represents a joined table for an endpoint.
//
// For Type see enums.JoinType.
//
// IncludeJoinFields, when true, indicates that the fields from the joined table,
// which are used for the join, are included in the output.
// By default, they are excluded.
//
// QualifiedFields, when true, indicates that for a sibling join
// all fields at the same level must be fully qualified.
//
// Target is the target table to be joined.
// If empty, the primary table from the endpoint will be used as the target.
//
// Source is the source table to join with the target table.
//
// Fields lists the field identifiers to use for the join.
// If empty, the join will be performed using a foreign-key in the joined table
// which matches the primary key of the primary table.
// If no foreign key matches, the join will fail!
//
// The optional JoinID is used for parent/child type joins.
// If it is not empty, it will be used as the identifier of the field containing the joined records.
// If it is empty, the identifier of the table is used.
// Make sure the definition resolves to a unique identifier,
// so it doesn't overwrite a field with the same code.
//
// Joins are optional subjoins, which will be executed against the source table of this join.
type Join struct {
	Type              enums.JoinType `json:"type"                        yaml:"type"`
	IncludeJoinFields bool           `json:"includeJoinFields,omitempty" yaml:"includeJoinFields,omitempty"`
	QualifiedFields   bool           `json:"qualifiedFields,omitempty"   yaml:"qualifiedFields,omitempty"`
	Target            string         `json:"target,omitempty"            yaml:"target,omitempty"`
	Source            string         `json:"source"                      yaml:"source"`
	Fields            []string       `json:"fields,omitempty"            yaml:"fields,omitempty"`
	JoinID            string         `json:"joinID,omitempty"            yaml:"joinID,omitempty"`
	Joins             []*Join        `json:"joins,omitempty"             yaml:"joins,omitempty"`
	// hidden fields
	mutex  sync.Mutex
	target *Table
	source *Table
}

// GetTarget returns the primary table definition.
func (j *Join) GetTarget() *Table {
	return j.target
}

// GetSource returns the joined table definition.
func (j *Join) GetSource() *Table {
	return j.source
}

// GetJoinID returns the id to use for parent/child relationships.
func (j *Join) GetJoinID() string {
	if j.JoinID != "" {
		return j.JoinID
	}
	return j.Source
}

// Fix (re)sets the parent-child relationships for this object.
func (j *Join) Fix(ds *Datasource) {
	j.mutex.Lock()
	j.fix(ds)
	j.mutex.Unlock()
}

func (j *Join) fix(ds *Datasource) {
	if ds != nil {
		j.source = ds.Table(j.Source)

		if j.Target != "" {
			j.target = ds.Table(j.Target)
		}
	}

	for _, j2 := range j.Joins {
		j2.Fix(ds)
	}
}
