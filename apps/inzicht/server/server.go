// Package server implements the HTTP service for the Inzicht app: aggregate
// usage statistics per doel/afnemer (shared with the verstrekker) and an Inzicht
// API through which the verstrekker requests a set of processing activities,
// gated by an approval step on the afnemer side.
//
// The service is deliberately lean: it uses the standard library net/http server
// (Go 1.22+ method-aware routing) rather than the full EAM/fiber stack, while
// reusing the shared EAM configuration, the ADL query layer (eam/log/adl/query)
// and the Valkeyrie KV persistence pattern.
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/inzicht/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/inzicht/verzoek"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl/query"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// Service is the Inzicht HTTP service.
type Service struct {
	cfg    *config.Config
	logger *slog.Logger
	ctx    context.Context

	store  *verzoek.Store
	adl    *adl.Logger
	source query.Source
	mux    *http.ServeMux

	// tokens is the bearer-token -> principal table (simulation stand-in for an
	// OAuth2/mTLS token service); authDisabled turns auth off for local play.
	tokens       map[string]*principal
	authDisabled bool
}

// NewService constructs the Inzicht service: it wires the Verzoek KV store
// (in-memory by default, Postgres/etcd/consul as configured), the ADL logger
// (for logging access to the Inzicht API itself, into the same write-ahead log),
// the ADL query source over the write-ahead log, and the HTTP routes.
func NewService(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Service, error) {
	// Persistence: reuse the shared Valkeyrie KV pattern. When no backend type is
	// configured (or "memory"), fall back to the in-memory store, mirroring the PAP.
	var kv store.Store
	if cfg.Persist.Type != "" {
		var err error
		if kv, err = cfg.Persist.NewStore(ctx); err != nil {
			return nil, fmt.Errorf("inzicht: create persistence store: %w", err)
		}
	}
	if kv == nil {
		kv = memory.New()
	}

	base := cfg.Persist.Base
	if base == "" {
		base = "inzicht/verzoeken"
	}

	logRec, err := adl.New(adl.Config{
		Level:    cfg.ADL.Level,
		WALPath:  cfg.ADL.Path,
		Resource: cfg.ADL.ResourceMap(map[string]any{"service.name": "inzicht"}),
		Logger:   logger,
	})
	if err != nil {
		return nil, fmt.Errorf("inzicht: create ADL logger: %w", err)
	}

	source, err := buildSource(cfg)
	if err != nil {
		return nil, err
	}

	s := &Service{
		cfg:          cfg,
		logger:       logger,
		ctx:          ctx,
		store:        verzoek.NewStore(ctx, kv, base),
		adl:          logRec,
		source:       source,
		tokens:       parseAuthTokens(cfg.Inzicht.AuthTokens),
		authDisabled: cfg.Inzicht.AuthDisabled,
	}
	logger.Info("inzicht: read-side backend selected", "backend", backendName(cfg.Inzicht.Backend))
	if s.authDisabled {
		logger.Warn("inzicht: API authentication is DISABLED (INZICHT_AUTH_DISABLED) - local play only")
	} else if len(s.tokens) == 0 {
		logger.Warn("inzicht: no bearer tokens configured (INZICHT_AUTH_TOKENS) - the API is closed (401)")
	}
	s.mux = s.routes()
	return s, nil
}

// Handler returns the HTTP handler (useful for tests via httptest). It wraps the
// routing mux in the authentication middleware so every protected route requires
// an authenticated caller.
func (s *Service) Handler() http.Handler { return s.authenticate(s.mux) }

// Serve starts the HTTP server and, when configured, the periodic statistics
// push. It blocks until the server stops.
func (s *Service) Serve() error {
	s.StartPush(s.ctx)

	addr := net.JoinHostPort(s.cfg.Server.Host, fmt.Sprintf("%d", s.cfg.Server.Port))
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.authenticate(s.mux),
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
		IdleTimeout:  s.cfg.Server.IdleTimeout,
	}

	if s.cfg.Server.Cert != "" && s.cfg.Server.Key != "" {
		srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		s.logger.Info("inzicht service listening (https)", "addr", addr)
		return srv.ListenAndServeTLS(s.cfg.Server.Cert, s.cfg.Server.Key)
	}
	s.logger.Info("inzicht service listening (http)", "addr", addr)
	return srv.ListenAndServe()
}

// Close releases resources held by the service.
func (s *Service) Close() error {
	if s.adl != nil {
		return s.adl.Close()
	}
	return nil
}

// bucket resolves the configured/queried period bucket.
func (s *Service) bucket(q string) query.Bucket {
	if q == "" {
		q = s.cfg.Inzicht.Bucket
	}
	switch query.Bucket(q) {
	case query.BucketWeek:
		return query.BucketWeek
	case query.BucketMonth:
		return query.BucketMonth
	default:
		return query.BucketDay
	}
}

// kThreshold returns the configured k-anonymity threshold (>=1).
func (s *Service) kThreshold() int {
	if s.cfg.Inzicht.K > 0 {
		return s.cfg.Inzicht.K
	}
	return query.DefaultK
}

// now is overridable in tests.
var now = time.Now
