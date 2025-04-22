package server

import (
	"fmt"

	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

func (s *service) newPIP() pip2.PIP {
	var opts []pip2.Option

	if s.store != nil {
		base := fmt.Sprintf(convert.ForceSuffix(s.cfg.Persist.Base, pip2.PathSeparator), "data", pip2.PathSeparator)
		opts = append(opts, pip2.WithPersistence(s.store, base))
	}

	return pip2.New(s.ctx, s.logger, opts...)
}
