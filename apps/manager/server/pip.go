package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

func (s *service) newPIP() pip.PIP {
	var opts []pip.Option

	if s.store != nil {
		base := fmt.Sprintf(convert.ForceSuffix(s.cfg.Persist.Base, pip.PathSeparator), "data", pip.PathSeparator)
		opts = append(opts, pip.WithPersistence(s.store, base))
	}

	return pip.New(s.ctx, s.logger, opts...)
}
