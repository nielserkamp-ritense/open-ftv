// Package principals stores the manager's local record of every Principal it has seen act on
// the management plane, so an attribution string (created_by, updated_by, *_audit.user_id) can
// be rendered as a person and linked to.
//
// The stored record is display-facing only. It is never an input to an authorization decision:
// roles stay in the token and the PDP remains authoritative. See docs/adr/0002 and docs/adr/0004.
package principals

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
)

// Kind values as stored in principal.kind. These are storage-level values rather than
// identity.Kind: KindLegacy and KindSystem rows are written by migrations and seeds, never by a
// request, so identity.Kind deliberately has no counterpart for them.
const (
	// KindUser is a subject the manager has seen authenticate against an IdP.
	KindUser = "user"
	// KindSystem is one of the manager's own *SYSTEM*-style attribution sentinels.
	KindSystem = "system"
	// KindLegacy is an attribution value that predates the principal table: it is keyed by a
	// display name rather than an id, so it can be shown but must not be linked to.
	KindLegacy = "legacy"
)

// Record is the stored, display-facing view of a Principal.
//
// Name is empty when the manager knows the id but has not yet seen its owner authenticate — a
// row carried over from the audit tables. Callers fall back to showing the raw id.
type Record struct {
	ID    string
	Kind  string
	Name  string
	Email string
}

// Linkable reports whether the UI can link to r, which requires the record to identify a real
// subject. A legacy record is keyed by a display name and identifies nobody; a system record is
// the manager itself.
func (r Record) Linkable() bool {
	return r.Kind == KindUser
}

// Store records the Principals the manager has seen, and resolves stored attribution ids back
// into display data.
type Store interface {
	// Upsert records that p was seen acting on the management plane.
	Upsert(ctx context.Context, p identity.Principal) error
	// Resolve returns a Record for each of the given ids that is known. Ids with no record are
	// absent from the map: the caller falls back to showing the raw id.
	Resolve(ctx context.Context, ids []string) (map[string]Record, error)
}
