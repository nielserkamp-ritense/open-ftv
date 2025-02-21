package server

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pap/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pap/persistence"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// NewPAP instantiates a vanilla PAP with an optional persistent storage backend.
func NewPAP(ctx context.Context, cfg *config.Config, logger *slog.Logger) (pap.PAP, error) {
	l := models.LanguageFromString(cfg.PolicyLanguage)
	opts := []pap.Option{pap.WithLanguage(l.Language()), pap.WithFileStore(cfg.PolicyStore, cfg.PolicyStoreRecurse)}

	if cfg.PersistType != "" {
		s, err := persistence.New(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if s != nil {
			opts = append(opts, pap.WithPersistence(s, cfg.PersistBase))
		}
	}

	return pap.New(ctx, logger, opts...), nil
}
