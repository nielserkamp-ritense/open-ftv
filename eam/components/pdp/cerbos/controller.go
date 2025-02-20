// Package cerbos contains all logic for a functional component acting as the Policy Decision Point
// using Cerbos/CEL as the policy language.
package cerbos

import (
	"errors"
	"log/slog"
	"path/filepath"

	"github.com/cerbos/cerbos-sdk-go/cerbos"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
)

// Version defines the version of this Cerbos/CEL PDP.
const Version = "1.0.0"

// Config defines the parameters to configure a new Cerbos/CEL PDP.
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

	c := &controller{Base: pdp.NewBase(options...), cfg: cfg, policyIDs: make(map[string]string)}
	c.logger = c.Logger().With("controller", c.String(), "clientAddress", c.cfg.Addr1, "adminAddress", c.cfg.Addr2, "ca", c.cfg.CA)

	if c.PAP() != nil {
		c.PAP().AddEventSink(c)
	}

	c.initClients()

	if store, recurse := c.Store(); store != "" {
		store, _ = filepath.Abs(store)
		c.PAP().LoadFromStore(store, recurse)
	}

	c.logger.Info("pbac controller initialized")
	return c
}

func (c *controller) initClients() {
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
		c.logger.Error("failed to initialize clients", "error", errors.Join(err, err2))
	case err != nil:
		c.logger.Error("failed to initialize gRPC client", "error", err)
	case err2 != nil:
		c.logger.Error("failed to initialize admin client", "error", err2)
	default:
		if c.info, err = c.engine.ServerInfo(c.Context()); err != nil || c.info == nil {
			c.logger.Error("failed to retrieve server-info", "error", err)
		} else {
			c.logger = c.logger.With("serverInfo", c.info)
		}
	}
}

type controller struct {
	pdp.Base
	cfg       Config
	logger    *slog.Logger
	admin     *cerbos.GRPCAdminClient
	engine    *cerbos.GRPCClient
	info      *cerbos.ServerInfo
	policyIDs map[string]string
}
