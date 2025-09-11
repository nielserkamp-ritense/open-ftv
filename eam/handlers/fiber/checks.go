package fiber

import (
	"errors"
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

func newFieldChecker() *fieldChecker {
	return &fieldChecker{errs: make([]error, 0)}
}

type fieldChecker struct {
	prefix string
	errs   []error
}

func (f *fieldChecker) checkFailed() bool {
	return len(f.errs) > 0
}

func (f *fieldChecker) error() error {
	return errors.Join(f.errs...)
}

func (f *fieldChecker) add(msg string) {
	if f.prefix != "" {
		msg = fmt.Sprintf("%s %s", f.prefix, msg)
	}
	f.errs = append(f.errs, errors.New(msg))
}

func (f *fieldChecker) checkIdentifiers(id1 string, id2 *string, msg string) *fieldChecker {
	if id1 != *id2 {
		if *id2 != "" {
			f.add(msg)
		} else {
			*id2 = id1
		}
	}
	return f
}

func (f *fieldChecker) checkLanguage(language string) *fieldChecker {
	switch {
	case language == "":
		f.add("language must be filled")
	case len(language) > 40:
		f.add("language too long (max 40 characters)")
	}
	return f
}

func (f *fieldChecker) checkTitle(title string) *fieldChecker {
	switch {
	case title == "":
		f.add("title must be filled")
	case len(title) > 80:
		f.add("title too long (max 80 characters)")
	}
	return f
}

func (f *fieldChecker) checkRvvaID(id string) *fieldChecker {
	switch {
	case len(id) > 80:
		f.add("RvVA identifier too long (max 40 characters)")
	}
	return f
}

func (f *fieldChecker) checkAttrType(tp string) *fieldChecker {
	switch {
	case len(tp) > 40:
		f.add("attribute type too long (max 40 characters)")
	}
	return f
}

func (f *fieldChecker) checkEntityType(tp string) *fieldChecker {
	switch {
	case len(tp) > 80:
		f.add("entity type too long (max 80 characters)")
	}
	return f
}

func (f *fieldChecker) checkPolicyData(url, data string) *fieldChecker {
	switch {
	case url == "" && data == "":
		f.add("either uri or data must be filled")
	case url != "" && data != "":
		f.add("only one of uri and data can be filled")
	case len(url) > 400:
		f.add("uri too long (max 400 characters)")
	}
	return f
}

func (f *fieldChecker) checkTags(tags []string) *fieldChecker {
	for i := range tags {
		tag := tags[i]
		switch {
		case tag == "":
			f.add(fmt.Sprintf("tag #%d must not be empty", i+1))
		case len(tag) > 40:
			f.add(fmt.Sprintf("tag #%d is too long (max 40 characters)", i+1))
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
			f.add("key must be filled")
		case len(a.Key) > 200:
			f.add("key too long (max 200 characters)")
		}

		f.checkAttrType(a.Type).checkTitle(a.Metadata.Title)
	}

	f.prefix = ""
	return f
}
