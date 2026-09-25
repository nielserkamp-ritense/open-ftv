package config

// Cerbos contains the configuration variables for connecting with Cerbos PDPs.
type Cerbos struct {
	Address      string `env:"CERBOS_ADDRESS" flag:"cerbos-address" desc:"Address of the Cerbos client interface"`
	CA           string `env:"CERBOS_CA" flag:"cerbos-ca" desc:"File containing the CA certificate for the Cerbos interfaces"`
	AdminAddress string `env:"CERBOS_ADMIN" flag:"cerbos-admin" desc:"Address of the Cerbos admin interface"`
	User         string `env:"CERBOS_USER" flag:"cerbos-user" desc:"User-id for the Cerbos admin interface"`
	Pswd         string `env:"CERBOS_PSWD" flag:"cerbos-pswd" desc:"Password for the Cerbos admin interface"`
}
