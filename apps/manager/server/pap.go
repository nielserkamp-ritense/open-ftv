package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// NewPAP instantiates a vanilla PAP with an optional persistent storage backend.
func (s *service) newPAP() *pap.PAP {
	opts := []pap.Option{pap.WithLanguage(s.l.Language())}

	if s.store != nil {
		base := fmt.Sprintf("%s%s%s", convert.ForceSuffix(s.cfg.Persist.Base, pap.PathSeparator), "policies", pap.PathSeparator)
		opts = append(opts, pap.WithPersistence(s.store, base))
	}

	return pap.New(s.ctx, s.logger, opts...)
}
