package server

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// NewPAP instantiates a vanilla PAP with an optional persistent storage backend.
func (s *service) newPAP() (*pap.PAP, error) {
	opts := []pap.Option{pap.WithLanguage(s.l.Language())}
	isPG := strings.EqualFold(s.cfg.Persist.Type, "postgres")

	switch {
	case isPG:
		db, err := pap.NewPostgresDB(s.ctx, s.cfg.Persist.PgURL, s.cfg.Persist.PgMaxLife, s.cfg.PgMaxConn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect postgres backend: %w", err)
		}
		opts = append(opts, pap.WithPostgresDB(db))

	case s.store != nil:
		base := fmt.Sprintf("%s%s%s", convert.ForceSuffix(s.cfg.Persist.Base, pap.PathSeparator), "policies", pap.PathSeparator)
		opts = append(opts, pap.WithKeyValueDB(s.store, base))
	}

	if s.cfg.Migration.Source != "" {
		switch {
		case isPG:
			opts = append(opts, pap.WithMigration(s.cfg.Migration.Source, s.cfg.Persist.PgURL, s.cfg.Migration.Auto, s.cfg.Migration.Steps))
		}
	}

	p := pap.New(s.ctx, s.logger, opts...)
	if p == nil {
		return nil, fmt.Errorf("failed to create pap")
	}
	return p, nil
}
