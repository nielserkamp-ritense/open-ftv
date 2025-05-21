package schema

import "fmt"

// Parent represents the parent in a parent-child relationship.
type Parent struct {
	ID string `json:"id" yaml:"id"`
	// hidden fields
	parent *Parent
}

// FQDN returns the fully qualified ID of this parent.
func (p *Parent) FQDN() string {
	if p.parent != nil {
		return fmt.Sprintf("%s.%s", p.parent.FQDN(), p.ID)
	}
	return p.ID
}
