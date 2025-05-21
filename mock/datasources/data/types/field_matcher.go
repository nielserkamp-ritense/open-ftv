package types

import (
	"fmt"
	"regexp"
	"strings"
)

// FieldMatcher represents the interface to test if a field identifier matches a given expression.
type FieldMatcher interface {
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
			out.always = true
			return out
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

// Always implements the FieldMatcher interface.
func (m *matcher) Always() bool {
	return m.always
}

// Match implements the FieldMatcher interface.
func (m *matcher) Match(id string) bool {
	if m.always {
		return true
	}

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
	".", "",
	"+", "",
	"{", "",
	"}", "",
	"\\", "",
	"*", ".*",
	"?", ".",
)
