// Package persistence handles persistent storage for the PAP.
package persistence

import (
	"context"
	"strings"

	"github.com/kvtools/consul"
	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pip/config"
)

// newConsul instantiates persistent storage with an etcd backend.
func newConsul(ctx context.Context, cfg *config.Config) (store.Store, error) {
	addr := strings.Split(cfg.PersistAddresses, ",")
	return valkeyrie.NewStore(ctx, consul.StoreName, addr, &consul.Config{
		ConnectionTimeout: cfg.PersistTimeout,
		Token:             cfg.ConsulToken,
		Namespace:         cfg.ConsulNamespace,
	})
}
