package server

import (
	"fmt"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// NewPAP instantiates a vanilla PAP with an optional persistent storage backend.
func (s *service) newPAP() pap2.PAP {
	opts := []pap2.Option{pap2.WithLanguage(s.l.Language())}

	if s.store != nil {
		base := fmt.Sprintf("%s%s%s", convert.ForceSuffix(s.cfg.Persist.Base, pap2.PathSeparator), "policies", pap2.PathSeparator)
		opts = append(opts, pap2.WithPersistence(s.store, base))
	}

	return pap2.New(s.ctx, s.logger, opts...)
}
