package matching

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// FieldMatcher represents the interface to test if a field identifier matches a given expression.
type FieldMatcher interface {
	String() string
	Always() bool
	Match(id string) bool
}

// NewFieldMatcher instantiates a new set of field matchers.
func NewFieldMatcher(exp string) FieldMatcher {
	if exp == "" {
		return &matcher{always: true}
	}

	parts := strings.Split(strings.ToLower(exp), ",")
	out := &matcher{}

	for i := range parts {
		part := parts[i]
		switch {
		case part == "":
			// ignore
		case part == "*":
			return &matcher{always: true}
		case strings.Contains(part, "*") || strings.Contains(part, "?"):
			expr := makeExpr.Replace(part)
			out.rx = append(out.rx, regexp.MustCompile(fmt.Sprintf("^%s$", expr)))
		default:
			out.exact = append(out.exact, part)
		}
	}

	return out
}

type matcher struct {
	always bool
	exact  []string
	rx     []*regexp.Regexp
}

// String implements the Stringer interface.
func (m *matcher) String() string {
	b := bytes.Buffer{}
	b.Grow(len(m.exact)*20 + len(m.rx)*20)

	b.WriteByte('{')

	if m.always {
		b.WriteString("always:true")
	} else {
		if len(m.exact) > 0 {
			b.WriteString("exact:[")
			for i := range m.exact {
				if i > 0 {
					b.WriteByte(',')
				}
				b.WriteString(m.exact[i])
			}
			b.WriteByte(']')
		}

		if len(m.rx) > 0 {
			if len(m.exact) > 0 {
				b.WriteByte(',')
			}

			b.WriteString("regex:[")
			for i := range m.rx {
				if i > 0 {
					b.WriteByte(',')
				}
				b.WriteString(m.rx[i].String())
			}
			b.WriteByte(']')
		}
	}

	b.WriteByte('}')
	return b.String()
}

// Always implements the FieldMatcher interface.
func (m *matcher) Always() bool {
	return m.always
}

// Match implements the FieldMatcher interface.
func (m *matcher) Match(id string) bool {
	if m.always {
		return true
	}

	id = strings.ToLower(id)

	for i := range m.exact {
		if id == m.exact[i] {
			return true
		}
	}

	for _, rx := range m.rx {
		if rx.MatchString(id) {
			return true
		}
	}

	return false
}

var makeExpr = strings.NewReplacer(
	"(", "",
	")", "",
	"[", "",
	"]", "",
	".", "\\.",
	"+", "",
	"{", "",
	"}", "",
	"\\", "",
	"*", ".*",
	"?", ".",
)
