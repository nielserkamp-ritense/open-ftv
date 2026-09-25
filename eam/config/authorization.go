package config

// Authorization contains the configuration variables for API endpoints to authorize access for users and/or external processes.
type Authorization struct {
	Authenticate bool `json:"forceAuthentication" yaml:"authorization.authenticate,omitempty" env:"AUTHORIZATION_AUTHENTICATE" flag:"authorization-authenticate" desc:"Indicates the authorization process must first authenticate the user an/or process"`
	// FailClosedOnEmpty makes an empty or unreadable policy store DENY all requests
	// instead of allowing them. Default false preserves the legacy fail-open behavior
	// (no policies => NoAuth => allow all). The management plane (manager) enables this
	// so a fresh/wiped/unreadable store locks down rather than exposing the API.
	FailClosedOnEmpty bool `json:"failClosedOnEmpty,omitempty" yaml:"authorization.fail_closed_on_empty,omitempty" env:"AUTHORIZATION_FAIL_CLOSED_ON_EMPTY" flag:"authorization-fail-closed-on-empty" desc:"Deny all requests when the policy store is empty or unreadable, instead of allowing them (fail closed)"`
}
