package config

import "flag"

// OIDC configures in-process JWT validation for an app's PEP.
type OIDC struct {
	Issuer     string
	JWKSURL    string
	Audience   string
	RolesClaim string
}

func (o *OIDC) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.Issuer, "oidc-issuer", "", "Expected issuer of bearer tokens")
	fs.StringVar(&o.JWKSURL, "oidc-jwks-url", "", "URL of the JWKS for validating bearer tokens")
	fs.StringVar(&o.Audience, "oidc-audience", "", "Expected audience of bearer tokens")
	fs.StringVar(&o.RolesClaim, "oidc-roles-claim", "", "Claim in bearer tokens holding the roles")
}
