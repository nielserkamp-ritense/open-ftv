// Package cerbos contains all logic for a functional component acting as the Policy Decision Point
// using Cerbos/CEL as the policy language.
package cerbos

import (
	"errors"
	"path/filepath"

	"github.com/cerbos/cerbos-sdk-go/cerbos"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
)

// Version defines the version of this Cerbos/CEL PDP.
const Version = "1.0.0"

type Config struct {
	Addr1 string
	Addr2 string
	CA    string
	User  string
	Pswd  string
}

// NewController instantiates a new Cerbos/CEL controller.
func NewController(cfg Config, options ...pdp.Option) pdp.Controller {
	options = append(options, pdp.WithNameVersion(components.CERBOS.String(), Version))

	c := &controller{Base: pdp.NewBase(options...), cfg: cfg}
	c.SetPAP(pap.New(c.Context(), c.Logger(), c))

	var err, err2 error

	if c.cfg.CA == "" {
		c.cfg.CA = "<noTLS>"
		c.engine, err = cerbos.New(c.cfg.Addr1, cerbos.WithPlaintext())
		c.admin, err2 = cerbos.NewAdminClientWithCredentials(c.cfg.Addr2, c.cfg.User, c.cfg.Pswd, cerbos.WithPlaintext())
	} else {
		c.engine, err = cerbos.New(c.cfg.Addr1, cerbos.WithTLSCACert(c.cfg.CA), cerbos.WithTLSInsecure())
		c.admin, err2 = cerbos.NewAdminClientWithCredentials(c.cfg.Addr2, c.cfg.User, c.cfg.Pswd, cerbos.WithTLSCACert(c.cfg.CA), cerbos.WithTLSInsecure())
	}

	switch {
	case err != nil && err2 != nil:
		c.Logger().Error("failed to create clients", c.log(errors.Join(err, err2))...)
	case err != nil:
		c.Logger().Error("failed to create gRPC client", c.log(err)...)
	case err2 != nil:
		c.Logger().Error("failed to create admin client", c.log(err2)...)
	default:
		if c.info, err = c.engine.ServerInfo(c.Context()); err != nil || c.info == nil {
			c.Logger().Error("failed to retrieve server-info", c.log(err)...)
		}
	}

	if store, recurse := c.Store(); store != "" {
		store, _ = filepath.Abs(store)
		c.PAP().LoadFromStore(store, recurse)
	}

	c.Logger().Info("pbac controller initialized", c.log(nil)...)
	return c
}

func (c *controller) log(err error) []any {
	out := []any{"controller", c.String(), "clientAddress", c.cfg.Addr1, "adminAddress", c.cfg.Addr2, "ca", c.cfg.CA, "serverInfo", c.info}
	if err != nil {
		out = append(out, "error", err)
	}
	return out
}

type controller struct {
	pdp.Base
	cfg    Config
	admin  *cerbos.GRPCAdminClient
	engine *cerbos.GRPCClient
	info   *cerbos.ServerInfo
}
