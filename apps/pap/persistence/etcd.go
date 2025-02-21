package persistence

import (
	"context"
	"strings"

	"github.com/kvtools/etcdv3"
	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pap/config"
)

// newETCD instantiates persistent storage with an etcd backend.
func newETCD(ctx context.Context, cfg *config.Config) (store.Store, error) {
	addr := strings.Split(cfg.PersistAddresses, ",")
	return valkeyrie.NewStore(ctx, etcdv3.StoreName, addr, &etcdv3.Config{
		ConnectionTimeout: cfg.PersistTimeout,
		SyncPeriod:        cfg.EtcdSync,
		Username:          cfg.EtcdUser,
		Password:          cfg.EtcdPswd,
	})
}
