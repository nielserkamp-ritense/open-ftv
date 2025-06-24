package config

// Cerbos contains the configuration variables for connecting with Cerbos PDPs.
type Cerbos struct {
	Address      string `yaml:"cerbos.address,omitempty" env:"CERBOS_ADDRESS" flag:"cerbos-address" desc:"Address of the Cerbos client interface"`
	CA           string `yaml:"cerbos.ca,omitempty" env:"CERBOS_CA" flag:"cerbos-ca" desc:"File containing the CA certificate for the Cerbos interfaces"`
	AdminAddress string `yaml:"cerbos.admin.address,omitempty" env:"CERBOS_ADMIN" flag:"cerbos-admin" desc:"Address of the Cerbos admin interface"`
	User         string `yaml:"cerbos.admin.user,omitempty" env:"CERBOS_USER" flag:"cerbos-user" desc:"User-id for the Cerbos admin interface"`
	Pswd         string `yaml:"cerbos.admin.password,omitempty" env:"CERBOS_PSWD" flag:"cerbos-pswd" desc:"Password for the Cerbos admin interface"`
}

// Sanitized returns the configuration variables where all sensitive data has been scrubbed.
func (c *Cerbos) Sanitized() *Cerbos {
	sanitized := *c
	sanitized.User = ""
	sanitized.Pswd = ""
	return &sanitized
}
