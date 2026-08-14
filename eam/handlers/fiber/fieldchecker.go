package fiber

import (
	"errors"
	"fmt"
)

type checkIssue struct {
	code string
	msg  string
}

type fieldChecker struct {
	prefix string
	errs   []error
}

func newFieldChecker() *fieldChecker {
	return &fieldChecker{errs: make([]error, 0)}
}

func (f *fieldChecker) checkFailed() bool {
	return len(f.errs) > 0
}

func (f *fieldChecker) error() error {
	return errors.Join(f.errs...)
}

// firstCode returns the Code of the first recorded check failure, or "" if none had one.
func (f *fieldChecker) firstCode() string {
	if len(f.errs) == 0 {
		return ""
	}

	if e, ok := f.errs[0].(*checkErr); ok {
		return e.code
	}

	return ""
}

// addIssue records a check failure from a checkIssue.
func (f *fieldChecker) addIssue(issue checkIssue) {
	f.addCode(issue.code, issue.msg)
}

// addIssuef records a check failure from a checkIssue whose Msg is a fmt.Sprintf template, rendered here with args.
func (f *fieldChecker) addIssuef(issue checkIssue, args ...any) {
	f.addCode(issue.code, fmt.Sprintf(issue.msg, args...))
}

// addCode records a check failure with a stable, machine-readable Code alongside its message.
func (f *fieldChecker) addCode(code, msg string) {
	if f.prefix != "" {
		msg = fmt.Sprintf("%s %s", f.prefix, msg)
	}

	f.errs = append(f.errs, &checkErr{code: code, msg: msg})
}

func (f *fieldChecker) checkIdentifiers(id1 string, id2 *string, issue checkIssue) *fieldChecker {
	if id1 != *id2 {
		if *id2 != "" {
			f.addIssue(issue)
		} else {
			*id2 = id1
		}
	}

	return f
}

// checkErr is a field-check failure that carries an optional stable Code alongside its message.
type checkErr struct {
	code string
	msg  string
}

func (e *checkErr) Error() string {
	return e.msg
}
