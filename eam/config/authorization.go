package config

import "flag"

// Authorization contains the configuration variables for API endpoints to authorize access for users and/or external processes.
type Authorization struct {
	Authenticate bool
	// FailClosedOnEmpty makes an empty or unreadable policy store DENY all requests
	// instead of allowing them. Default false preserves the legacy fail-open behavior
	// (no policies => NoAuth => allow all). The management plane (manager) enables this
	// so a fresh/wiped/unreadable store locks down rather than exposing the API.
	FailClosedOnEmpty bool
}

func (a *Authorization) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&a.Authenticate, "authorization-authenticate", false, "Indicates the authorization process must first authenticate the user an/or process")
	fs.BoolVar(&a.FailClosedOnEmpty, "authorization-fail-closed-on-empty", false, "Deny all requests when the policy store is empty or unreadable, instead of allowing them (fail closed)")
}
