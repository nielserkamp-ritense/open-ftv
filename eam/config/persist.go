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

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql/pool"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/postgresql"
)

// Persist contains the configuration variables for various storage back-ends.
type Persist struct {
	Type            string        `json:"persistType,omitempty"      yaml:"persist.type,omitempty"                    env:"PERSIST_TYPE"              flag:"persist-type"              desc:"Persistence backend type (etcd, consul, postgres)"`
	Addresses       string        `json:"persistAddress,omitempty"   yaml:"persist.addresses,omitempty"               env:"PERSIST_ADDRESSES"         flag:"persist-addresses"         desc:"Persistence backend addresses"`
	Base            string        `json:"persistBase,omitempty"      yaml:"persist.prefix,omitempty"                  env:"PERSIST_PREFIX"            flag:"persist-prefix"            desc:"Persistence backend key prefix"`
	Timeout         time.Duration `json:"persistTimeout,omitempty"   yaml:"persist.timeout,omitempty"                 env:"PERSIST_TIMEOUT"           flag:"persist-timeout"           desc:"Persistence backend connection timeout"`
	EtcdSync        time.Duration `json:"persistEtcdSync,omitempty"  yaml:"persist.etcd.sync,omitempty"               env:"PERSIST_ETCD_SYNC"         flag:"persist-etcd-sync"         desc:"ETCD persistence backend sync period"`
	EtcdUser        string        `json:"-"                          yaml:"persist.etcd.user,omitempty"               env:"PERSIST_ETCD_USER"         flag:"persist-etcd-user"         desc:"ETCD persistence backend user"`
	EtcdPswd        string        `json:"-"                          yaml:"persist.etcd.password,omitempty"           env:"PERSIST_ETCD_PASSWORD"     flag:"persist-etcd-password"     desc:"ETCD persistence backend password"`
	ConsulToken     string        `json:"-"                          yaml:"persist.consul.token,omitempty"            env:"PERSIST_CONSUL_TOKEN"      flag:"persist-consul-token"      desc:"Consul persistence backend token"`
	ConsulNamespace string        `json:"persistConsulNS,omitempty"  yaml:"persist.consul.namespace,omitempty"        env:"PERSIST_CONSUL_NAMESPACE"  flag:"persist-consul-namespace"  desc:"Consul persistence backend namespace"`
	PgURL           string        `json:"-"                          yaml:"persist.postgres.url,omitempty"            env:"PERSIST_POSTGRES_URL"      flag:"persist-postgres-url"      desc:"Postgres persistence server url"`
	PgTable         string        `json:"persistPgTable,omitempty"   yaml:"persist.postgres.table,omitempty"          env:"PERSIST_POSTGRES_TABLE"    flag:"persist-postgres-table"    desc:"Postgres persistence database table"`
	PgMaxLife       time.Duration `json:"persistPgMaxLife,omitempty" yaml:"persist.postgres.connection.ttl,omitempty" env:"PERSIST_POSTGRES_CONN_TTL" flag:"persist-postgres-conn-ttl" desc:"Postgres persistence inactive connections time-to-live" default:"5m"`
	PgMaxConn       int32         `json:"persistPgMaxConn,omitempty" yaml:"persist.postgres.connection.max,omitempty" env:"PERSIST_POSTGRES_CONN_MAX" flag:"persist-postgres-conn-max" desc:"Postgres persistence maximum connections"               default:"100"`
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

// Validate reports whether Type names a supported persistence backend, so a misconfiguration can be
// caught before anything is constructed.
func (p *Persist) Validate() error {
	switch strings.ToLower(p.Type) {
	case "postgres", "postgresql", "pg", "etcd", "etcdv3", "consul": //nolint:goconst // type names, not worth centralizing
		return nil
	case "":
		return fmt.Errorf("persist-type is required: use one of postgres, etcd, or consul")
	default:
		return fmt.Errorf("unsupported persist-type %q: use one of postgres, etcd, or consul", p.Type)
	}
}

// NewStore returns a new persistence store based on the configuration variables.
//
// The function may produce an error if the configuration is invalid.
//
// The given context must be long-lived, and should be used to signal the store to shut down cleanly.
func (p *Persist) NewStore(ctx context.Context) (store.Store, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	switch strings.ToLower(p.Type) {
	case "postgres", "postgresql", "pg":
		return p.newPG(ctx)
	case "etcd", "etcdv3":
		return p.newETCD(ctx)
	default:
		return p.newConsul(ctx)
	}
}

func (p *Persist) newPG(ctx context.Context) (store.Store, error) {
	pool, err := pool.NewPool(
		ctx,
		p.PgURL,
		pool.WithMaxLifetime(p.PgMaxLife),
		pool.WithMaxConnections(p.PgMaxConn),
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
