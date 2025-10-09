// Package cerbos_api contains all logic for a functional component acting as the Policy Decision Point
// using Cerbos/CEL as the policy language.
package cerbos_api

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/cerbos/cerbos-sdk-go/cerbos"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
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
	options = append(options, pdp.WithNameVersion(models.CERBOS.String(), Version))

	c := &controller{Base: pdp.NewBase(options...), cfg: cfg, policyIDs: make(map[string]string)}
	c.Self = c
	c.logger = c.Logger.With("controller", c.String(), "clientAddress", c.cfg.Addr1, "adminAddress", c.cfg.Addr2, "ca", c.cfg.CA)

	c.initClients()

	if c.PAP != nil {
		c.PAP.AddEventSink(c)
		c.PAP.LoadFiles()
	}

	if c.ADL != nil {
		c.ADL.NewEngine(map[string]any{
			"controller":        c.Name,
			"controllerVersion": c.Version,
			"language":          models.CERBOS.String(),
			"clientAddress":     c.cfg.Addr1,
			"adminAddress":      c.cfg.Addr2,
			"serverInfo":        c.info,
		})
	}

	if c.engine != nil {
		if c.admin != nil {
			c.logger.Info("pdp controller initialized (with admin endpoint)")
		} else {
			c.logger.Info("pdp controller initialized (without admin endpoint)")
		}
	}
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
		if c.info, err = c.engine.ServerInfo(c.Ctx); err != nil || c.info == nil {
			c.logger.Error("failed to retrieve server-info", "error", err)
		} else {
			c.logger = c.logger.With("serverInfo", c.info)
		}
	}

	if c.engine != nil {
		// debug: include meta-data in the response
		c.engine = c.engine.With(cerbos.IncludeMeta(true))
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
	pdpMutex  sync.Mutex
}
