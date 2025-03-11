package server

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/manager/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/manager/persistence"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// NewPAP instantiates a vanilla PAP with an optional persistent storage backend.
func NewPAP(ctx context.Context, cfg *config.Config, logger *slog.Logger, l models.Language) (pap.PAP, error) {
	opts := []pap.Option{pap.WithLanguage(l.Language()), pap.WithFileStore(cfg.PolicyStore, cfg.PolicyStoreRecurse)}

	if cfg.PersistType != "" {
		s, err := persistence.New(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if s != nil {
			base := convert.ForceSuffix(cfg.PersistBase, "/") + "policies/"
			opts = append(opts, pap.WithPersistence(s, base))
		}
	}

	return pap.New(ctx, logger, opts...), nil
}
