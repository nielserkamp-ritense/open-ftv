package compare

import (
	"fmt"
	"regexp"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// RXFromLike builds a regular expression for a 'like' comparison.
func RXFromLike(in string) (*regexp.Regexp, error) {
	return buildRX(likeReplacer.Replace(convert.AnyToString(in)))
}

// RXFromString builds a regular expression for a 'regex' comparison.
func RXFromString(in string) (*regexp.Regexp, error) {
	return buildRX(convert.AnyToString(in))
}

func buildRX(s string) (*regexp.Regexp, error) {
	if !strings.HasPrefix(s, "^") {
		s = fmt.Sprintf("^%s", s)
	}
	if !strings.HasSuffix(s, "$") {
		s = fmt.Sprintf("%s$", s)
	}
	return regexp.Compile(s)
}

var likeReplacer = strings.NewReplacer(".", "\\.", "+", "\\+", "(", "\\(", ")", "\\)", "[", "\\[", "]", "\\]", "{", "\\{", "}", "\\}", "%", ".*", "*", ".*", "?", ".")
