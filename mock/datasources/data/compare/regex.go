package compare

import (
	"fmt"
	"regexp"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// RXFromLike builds a regular expression for a 'like' comparison.
func RXFromLike(in any) (*regexp.Regexp, error) {
	return buildRX(likeReplacer.Replace(convert.AnyToString(in)))
}

// RXFromString builds a regular expression for a 'regex' comparison.
func RXFromString(in any) (*regexp.Regexp, error) {
	return buildRX(convert.AnyToString(in))
}

func buildRX(in string) (*regexp.Regexp, error) {
	if !strings.HasPrefix(in, "^") {
		in = fmt.Sprintf("^%s", in)
	}
	if !strings.HasSuffix(in, "$") {
		in = fmt.Sprintf("%s$", in)
	}
	return regexp.Compile(in)
}

var likeReplacer = strings.NewReplacer(".", "\\.", "+", "\\+", "(", "\\(", ")", "\\)", "[", "\\[", "]", "\\]", "{", "\\{", "}", "\\}", "%", ".*", "*", ".*", "?", ".")
