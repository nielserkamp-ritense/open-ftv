package persistence

import (
	"context"
	"fmt"
	"strings"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pap/config"
)

// New instantiates a new persistent storage backend.
func New(ctx context.Context, cfg *config.Config) (store.Store, error) {
	switch strings.ToLower(cfg.PersistType) {
	case "memory", "mem":
		return nil, nil
	case "etcd", "etcdv3":
		return newETCD(ctx, cfg)
	case "consul":
		return newConsul(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported persistent storage type: %s", cfg.PersistType)
	}
}
