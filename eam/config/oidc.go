package config

import "flag"

// OIDC configures in-process JWT validation for an app's PEP.
type OIDC struct {
	Issuer     string `env:"OIDC_ISSUER"`
	JWKSURL    string `env:"OIDC_JWKS_URL"`
	Audience   string `env:"OIDC_AUDIENCE"`
	RolesClaim string `env:"OIDC_ROLES_CLAIM"`
}

func (o *OIDC) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.Issuer, "oidc-issuer", "", "Expected issuer of bearer tokens")
	fs.StringVar(&o.JWKSURL, "oidc-jwks-url", "", "URL of the JWKS for validating bearer tokens")
	fs.StringVar(&o.Audience, "oidc-audience", "", "Expected audience of bearer tokens")
	fs.StringVar(&o.RolesClaim, "oidc-roles-claim", "", "Claim in bearer tokens holding the roles")
}
