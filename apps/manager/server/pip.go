package server

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

func (s *Services) newPIP() (*pip.PIP, error) {
	var opts []pip.Option

	opts = append(opts, pip.WithPostgresDB(pip.NewPostgresWithPool(s.db)))

	return pip.New(s.ctx, s.logger, opts...)
}
