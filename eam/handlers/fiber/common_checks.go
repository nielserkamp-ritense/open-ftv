package fiber

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

const codeBadRequest = "E90400"

// Maximum lengths enforced by the field checks below.
const (
	maxStatusLength     = 40
	maxTitleLength      = 80
	maxAttrTypeLength   = 40
	maxEntityTypeLength = 80
	maxTagLength        = 40
	maxAttrKeyLength    = 200
)

var (
	errStatusTooLong     = checkIssue{code: "E95005", msg: "status too long (max 40 characters)"}
	errStatusInvalid     = checkIssue{code: "E95010", msg: "invalid status code"}
	errTitleTooLong      = checkIssue{code: "E95015", msg: "title too long (max 80 characters)"}
	errAttrTypeTooLong   = checkIssue{code: "E95020", msg: "attribute type too long (max 40 characters)"}
	errEntityTypeTooLong = checkIssue{code: "E95025", msg: "entity type too long (max 80 characters)"}
	errAttrKeyRequired   = checkIssue{code: "E95030", msg: "key must be filled"}
	errAttrKeyTooLong    = checkIssue{code: "E95035", msg: "key too long (max 200 characters)"}
	errTagEmptyFmt       = checkIssue{code: "E95040", msg: "tag #%d must not be empty"}
	errTagTooLongFmt     = checkIssue{code: "E95045", msg: "tag #%d is too long (max 40 characters)"}
)

func (f *fieldChecker) checkStatus(status string) *fieldChecker {
	switch {
	case status == "": // ignore
	case utf8.RuneCountInString(status) > maxStatusLength:
		f.addIssue(errStatusTooLong)
	case !strings.EqualFold(models.StatusFromString(status).String(), status):
		f.addIssue(errStatusInvalid)
	}

	return f
}

func (f *fieldChecker) checkTitle(title string) *fieldChecker {
	if utf8.RuneCountInString(title) > maxTitleLength {
		f.addIssue(errTitleTooLong)
	}

	return f
}

func (f *fieldChecker) checkAttrType(tp string) *fieldChecker {
	if utf8.RuneCountInString(tp) > maxAttrTypeLength {
		f.addIssue(errAttrTypeTooLong)
	}

	return f
}

func (f *fieldChecker) checkEntityType(tp string) *fieldChecker {
	if utf8.RuneCountInString(tp) > maxEntityTypeLength {
		f.addIssue(errEntityTypeTooLong)
	}

	return f
}

func (f *fieldChecker) checkTags(tags []string) *fieldChecker {
	for i := range tags {
		tag := tags[i]
		switch {
		case tag == "":
			f.addIssuef(errTagEmptyFmt, i+1)
		case utf8.RuneCountInString(tag) > maxTagLength:
			f.addIssuef(errTagTooLongFmt, i+1)
		}
	}

	return f
}

func (f *fieldChecker) checkAttributes(list []attributes.Attribute) *fieldChecker {
	for i := range list {
		a := &list[i]
		f.prefix = fmt.Sprintf("attribute #%d -", i+1)

		switch {
		case a.Key == "":
			f.addIssue(errAttrKeyRequired)
		case utf8.RuneCountInString(a.Key) > maxAttrKeyLength:
			f.addIssue(errAttrKeyTooLong)
		}

		f.checkAttrType(a.Type).checkTitle(a.Metadata.Title)
	}

	f.prefix = ""

	return f
}
