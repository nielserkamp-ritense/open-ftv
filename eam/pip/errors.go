package pip

import "errors"

// ErrInvalidStatusChange reports a status change the PIP data life cycle does not allow: an
// attribute or entity moves between concept and accepted only, in either direction.
var ErrInvalidStatusChange = errors.New("invalid status change")
