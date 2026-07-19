// Package identity holds the canonical representation of the request's caller, used for both PDP evaluation and audit attribution.
package identity

// Kind identifies the type of caller a Principal represents (user, app, system, ...).
type Kind struct {
	value string
}

// String returns k's underlying value, e.g. "user".
func (k Kind) String() string {
	return k.value
}

var (
	KindUser        = Kind{"user"}
	KindApp         = Kind{"app"}
	KindActivity    = Kind{"activity"}
	KindZaak        = Kind{"zaak"}
	KindDoelbinding = Kind{"doelbinding"}
	KindInvalid     = Kind{"invalid"}
	KindSystem      = Kind{"system"}
)

// Principal identifies the caller of a request, for both PDP evaluation and audit attribution (created_by/updated_by).
type Principal struct {
	Kind Kind
	ID   string
	// Name is the caller's display name (JWT preferred_username), used only for audit attribution.
	Name string
}

// DisplayName returns Name if set, otherwise ID — the value to persist into created_by/updated_by.
func (p Principal) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}

	return p.ID
}

// NewPrincipal builds a Principal of the given Kind and id.
func NewPrincipal(kind Kind, id string) Principal {
	return Principal{Kind: kind, ID: id}
}

// NewSystemPrincipal is the Principal used when no more specific identity applies.
func NewSystemPrincipal() Principal {
	return NewPrincipal(KindSystem, "*SYSTEM*")
}

// NewSeedPrincipal is the Principal attributed to policies seeded from disk at startup.
func NewSeedPrincipal() Principal {
	return NewPrincipal(KindSystem, "*SEED*")
}

// NewBootstrapPrincipal is the Principal attributed to the first, bootstrap bundle deployment.
func NewBootstrapPrincipal() Principal {
	return NewPrincipal(KindSystem, "*BOOTSTRAP*")
}

// NewStoragePrincipal is the Principal attributed to policy changes picked up from the storage watcher, e.g. edits made directly against the KV store.
// Note: as we will remove the KV store, this will eventually also be removed.
func NewStoragePrincipal() Principal {
	return NewPrincipal(KindSystem, "*STORAGE*")
}

// NewLoaderPrincipal is the Principal attributed to policies loaded from a local file-backed policy store at startup (PAP.LoadFiles).
func NewLoaderPrincipal() Principal {
	return NewPrincipal(KindSystem, "*LOADER*")
}

// NewBundlePrincipal is the Principal attributed to PAP/PIP state replaced by a bundle deployment.
func NewBundlePrincipal() Principal {
	return NewPrincipal(KindSystem, "*BUNDLE*")
}

// NewUnknownPrincipal is the Principal used when no caller has been resolved yet.
func NewUnknownPrincipal() Principal {
	return NewPrincipal(KindInvalid, "*UNKNOWN*")
}

// IsAuthenticatedUser reports whether p is an authenticated end user.
func (p Principal) IsAuthenticatedUser() bool {
	return p.Kind == KindUser && p.ID != ""
}

// String renders p as "kind::id", for display/logging only.
func (p Principal) String() string {
	return p.Kind.String() + "::" + p.ID
}

// Typed is satisfied by any value exposing a type and an ID (e.g. *models.Entity).
type Typed interface {
	Type() string
	ID() string
}

// FromEntity converts a Typed value into a Principal.
func FromEntity(e Typed) Principal {
	return NewPrincipal(Kind{e.Type()}, e.ID())
}
