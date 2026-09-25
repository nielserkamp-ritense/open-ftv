package config

import (
	"flag"

	cerbos_api "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
)

// Cerbos contains the configuration variables for connecting with Cerbos PDPs.
type Cerbos cerbos_api.Config

func (c *Cerbos) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Addr1, "cerbos-address", "", "Address of the Cerbos client interface")
	fs.StringVar(&c.CA, "cerbos-ca", "", "File containing the CA certificate for the Cerbos interfaces")
	fs.StringVar(&c.Addr2, "cerbos-admin", "", "Address of the Cerbos admin interface")
	fs.StringVar(&c.User, "cerbos-user", "", "User-id for the Cerbos admin interface")
	fs.StringVar(&c.Pswd, "cerbos-pswd", "", "Password for the Cerbos admin interface")
}
