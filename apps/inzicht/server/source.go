package server

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/inzicht/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl/query"
)

// backendName normalises the configured backend for logging ("" -> "wal").
func backendName(b string) string {
	if b = strings.ToLower(strings.TrimSpace(b)); b == "" {
		return "wal"
	}
	return b
}

// buildSource constructs the read-side query.Source selected by INZICHT_BACKEND. The
// statistics and Inzicht endpoints call Source.Query, so switching backend requires no
// change to the handlers - only the configuration. "wal" (default) reads the local
// write-ahead log; "loki" and "opensearch" serve from the observability backend.
func buildSource(cfg *config.Config) (query.Source, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Inzicht.Backend)) {
	case "", "wal":
		return query.NewWALSource(cfg.ADL.Path), nil

	case "loki":
		if cfg.Inzicht.LokiURL == "" {
			return nil, fmt.Errorf("inzicht: INZICHT_BACKEND=loki requires INZICHT_LOKI_URL")
		}
		return query.NewLokiSource(query.LokiConfig{
			URL:      cfg.Inzicht.LokiURL,
			OrgID:    cfg.Inzicht.LokiOrgID,
			User:     cfg.Inzicht.LokiUser,
			Password: cfg.Inzicht.LokiPassword,
		}), nil

	case "opensearch":
		if cfg.Inzicht.OpenSearchURL == "" {
			return nil, fmt.Errorf("inzicht: INZICHT_BACKEND=opensearch requires INZICHT_OPENSEARCH_URL")
		}
		return query.NewOpenSearchSource(query.OpenSearchConfig{
			URL:      cfg.Inzicht.OpenSearchURL,
			Index:    cfg.Inzicht.OpenSearchIndex,
			User:     cfg.Inzicht.OpenSearchUser,
			Password: cfg.Inzicht.OpenSearchPassword,
		}), nil

	default:
		return nil, fmt.Errorf("inzicht: unknown INZICHT_BACKEND %q (want wal, loki or opensearch)", cfg.Inzicht.Backend)
	}
}
