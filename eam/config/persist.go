package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kvtools/consul"
	"github.com/kvtools/etcdv3"
	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/postgresql"
)

// Persist contains the configuration variables for various storage back-ends.
type Persist struct {
	Type            string        `yaml:"persist.type,omitempty" env:"PERSIST_TYPE" flag:"persist-type" desc:"Persistence backend type (etcd, consul, postgres)"`
	Addresses       string        `yaml:"persist.addresses,omitempty" env:"PERSIST_ADDRESSES" flag:"persist-addresses" desc:"Persistence backend addresses"`
	Base            string        `yaml:"persist.prefix,omitempty" env:"PERSIST_PREFIX" flag:"persist-prefix" desc:"Persistence backend key prefix"`
	Timeout         time.Duration `yaml:"persist.timeout,omitempty" env:"PERSIST_TIMEOUT" flag:"persist-timeout" desc:"Persistence backend connection timeout"`
	EtcdSync        time.Duration `yaml:"persist.etcd.sync,omitempty" env:"PERSIST_ETCD_SYNC" flag:"persist-etcd-sync" desc:"ETCD persistence backend sync period"`
	EtcdUser        string        `yaml:"persist.etcd.user,omitempty" env:"PERSIST_ETCD_USER" flag:"persist-etcd-user" desc:"ETCD persistence backend user"`
	EtcdPswd        string        `yaml:"persist.etcd.password,omitempty" env:"PERSIST_ETCD_PASSWORD" flag:"persist-etcd-password" desc:"ETCD persistence backend password"`
	ConsulToken     string        `yaml:"persist.consul.token,omitempty" env:"PERSIST_CONSUL_TOKEN" flag:"persist-consul-token" desc:"Consul persistence backend token"`
	ConsulNamespace string        `yaml:"persist.consul.namespace,omitempty" env:"PERSIST_CONSUL_NAMESPACE" flag:"persist-consul-namespace" desc:"Consul persistence backend namespace"`
	PgURL           string        `yaml:"persist.postgres.url,omitempty" env:"PERSIST_POSTGRES_URL" flag:"persist-postgres-url" desc:"Postgres persistence server url"`
	PgTable         string        `yaml:"persist.postgres.table,omitempty" env:"PERSIST_POSTGRES_TABLE" flag:"persist-postgres-table" desc:"Postgres persistence database table"`
	PgMaxLife       time.Duration `yaml:"persist.postgres.connection.ttl,omitempty" env:"PERSIST_POSTGRES_CONN_TTL" flag:"persist-postgres-conn-ttl" default:"5m" desc:"Postgres persistence inactive connections time-to-live"`
	PgMaxConn       int32         `yaml:"persist.postgres.connection.max,omitempty" env:"PERSIST_POSTGRES_CONN_MAX" flag:"persist-postgres-conn-max" default:"100" desc:"Postgres persistence maximum connections"`
}

// Sanitized returns the configuration variables where all sensitive data has been scrubbed.
func (p *Persist) Sanitized() *Persist {
	sanitized := *p
	sanitized.EtcdUser = ""
	sanitized.EtcdPswd = ""
	sanitized.ConsulToken = ""
	sanitized.PgURL = ""
	return &sanitized
}

// NewStore returns a new persistence store based on the configuration variables.
//
// The function may produce an error if the configuration is invalid.
//
// The given context must be long-lived, and should be used to signal the store to shut down cleanly.
func (p *Persist) NewStore(ctx context.Context) (store.Store, error) {
	switch strings.ToLower(p.Type) {
	case "memory", "mem":
		return nil, nil // the default storage engine.
	case "postgres", "postgresql", "pg":
		return p.newPG(ctx)
	case "etcd", "etcdv3":
		return p.newETCD(ctx)
	case "consul":
		return p.newConsul(ctx)
	default:
		return nil, fmt.Errorf("unsupported persistent storage type: %s", p.Type)
	}
}

func (p *Persist) newPG(ctx context.Context) (store.Store, error) {
	pool, err := postgresql.NewPool(
		ctx,
		p.PgURL,
		postgresql.WithMaxLifetime(p.PgMaxLife),
		postgresql.WithMaxConnections(p.PgMaxConn),
	)
	if err != nil {
		return nil, err
	}

	return postgresql.New(pool, p.PgTable)
}

func (p *Persist) newETCD(ctx context.Context) (store.Store, error) {
	addr := strings.Split(p.Addresses, ",")
	return valkeyrie.NewStore(ctx, etcdv3.StoreName, addr, &etcdv3.Config{
		ConnectionTimeout: p.Timeout,
		SyncPeriod:        p.EtcdSync,
		Username:          p.EtcdUser,
		Password:          p.EtcdPswd,
	})
}

func (p *Persist) newConsul(ctx context.Context) (store.Store, error) {
	addr := strings.Split(p.Addresses, ",")
	return valkeyrie.NewStore(ctx, consul.StoreName, addr, &consul.Config{
		ConnectionTimeout: p.Timeout,
		Token:             p.ConsulToken,
		Namespace:         p.ConsulNamespace,
	})
}
