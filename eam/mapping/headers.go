package mapping

import (
	"slices"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// FromHeaders extracts a header value from the given headers attribute value.
//
// It is the plugin-facing helper for resolving a value from request headers:
// defaultKeys provides the header keys to look for, and any WithHeaderKeys option
// overrides them. Header keys are matched case-insensitively. It returns an empty
// string when none of the keys are present.
func FromHeaders(headers any, defaultKeys []string, opts ...Option) string {
	b := &base{headerKeys: defaultKeys}
	b.configure(opts)
	return b.fromHeaders(headers)
}

func (b *base) fromHeaders(headers any) string {
	// headers are case-insensitive, so we need to compare case-insensitive.
	var k string
	cmp := func(s string) bool { return strings.EqualFold(s, k) }

	// headers are supplied as the value of an attribute.
	// we need to determine the type.

	if m, ok := headers.(map[string]any); ok {
		for k = range m {
			if slices.ContainsFunc(b.headerKeys, cmp) {
				return convert.AnyToString(m[k])
			}
		}
	}

	if m, ok := headers.(map[string]string); ok {
		for k = range m {
			if slices.ContainsFunc(b.headerKeys, cmp) {
				return m[k]
			}
		}
	}

	return ""
}
