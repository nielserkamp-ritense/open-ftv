package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

func (s *Services) newPIP() *pip.PIP {
	var opts []pip.Option

	switch {
	case s.db != nil:
		opts = append(opts, pip.WithPostgresDB(pip.NewPostgresWithPool(s.db)))
	case s.store != nil:
		base := fmt.Sprintf(convert.ForceSuffix(s.cfg.Persist.Base, pip.PathSeparator), "data", pip.PathSeparator)
		opts = append(opts, pip.WithKeyValueDB(s.store, base))
	}

	return pip.New(s.ctx, s.logger, opts...)
}
