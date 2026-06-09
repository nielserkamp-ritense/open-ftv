package config

// OIDC configures in-process JWT validation for an app's PEP.
type OIDC struct {
	Issuer     string `env:"OIDC_ISSUER"`
	JWKSURL    string `env:"OIDC_JWKS_URL"`
	Audience   string `env:"OIDC_AUDIENCE"`
	RolesClaim string `env:"OIDC_ROLES_CLAIM"`
}
