package server

import (
	"fmt"

	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

func (s *service) newPIP() (pip2.PIP, error) {
	opts := make([]pip2.Option, 0)

	if s.cfg.Persist.Type != "" {
		store, err := s.cfg.Persist.NewStore(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if store != nil {
			opts = append(opts, pip2.WithPersistence(store, s.cfg.Persist.Base))
		}
	}

	// WARC logging of external information exchanges (shared with the push-ingest endpoint).
	w, err := s.cfg.PIP.NewWARC()
	if err != nil {
		return nil, fmt.Errorf("failed to create WARC log: %w", err)
	}
	if w != nil {
		s.warc = w
		opts = append(opts, pip2.WithWARC(w))
	}

	if s.cfg.PIP.PullConfigs != "" {
		opts = append(opts, pip2.WithPullConfigs(s.cfg.PIP.PullConfigs))
	}

	// emit an event for every attribute/entity mutation.
	opts = append(opts, pip2.WithEventSink(&slogEventSink{logger: s.logger}))

	return pip2.New(s.ctx, s.logger, opts...), nil
}
