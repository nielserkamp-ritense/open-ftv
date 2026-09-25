package config

import "flag"

// Cerbos contains the configuration variables for connecting with Cerbos PDPs.
type Cerbos struct {
	Address      string
	CA           string
	AdminAddress string
	User         string
	Pswd         string
}

func (c *Cerbos) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Address, "cerbos-address", "", "Address of the Cerbos client interface")
	fs.StringVar(&c.CA, "cerbos-ca", "", "File containing the CA certificate for the Cerbos interfaces")
	fs.StringVar(&c.AdminAddress, "cerbos-admin", "", "Address of the Cerbos admin interface")
	fs.StringVar(&c.User, "cerbos-user", "", "User-id for the Cerbos admin interface")
	fs.StringVar(&c.Pswd, "cerbos-pswd", "", "Password for the Cerbos admin interface")
}
