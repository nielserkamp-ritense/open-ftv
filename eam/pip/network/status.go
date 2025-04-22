package network

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// StatusCode represents the decoder parameters for a (range of) status code(s).
type StatusCode struct {
	Code        string             `json:"code" yaml:"code" toml:"code"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Attributes  *AttributesMapping `json:"attributes,omitempty" yaml:"attributes,omitempty" toml:"attributes,omitempty"`
	Entity      *EntityMapping     `json:"entity,omitempty" yaml:"entity,omitempty" toml:"entity,omitempty"`
	Relation    *RelationMapping   `json:"relation,omitempty" yaml:"relation,omitempty" toml:"relation,omitempty"`
	// hidden fields
	once   sync.Once
	codeRX *regexp.Regexp
}

// Match returns true if the given status code matches the code.
//
// An 'X' or 'x' character in the code matches any character in the status code.
// E.g. "2xx" matches any status code in the range [200..299].
func (s *StatusCode) Match(status int) bool {
	s.once.Do(func() {
		rx := strings.Replace(strings.ToLower(s.Code), "x", ".", -1)
		s.codeRX = regexp.MustCompile("^" + rx + "$")
	})

	str := fmt.Sprintf("%03d", status)
	return s.codeRX.MatchString(str)
}
