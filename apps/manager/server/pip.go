package server

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

func (s *Services) newPIP() (*pip.PIP, error) {
	return pip.New(s.ctx, s.logger, pip.WithPostgresDB(pip.NewPostgresWithPool(s.db)))
}
