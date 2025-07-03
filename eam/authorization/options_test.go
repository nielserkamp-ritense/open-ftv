package authorization

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	authentication2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelError)
	log := slog.New(h)

	p := pip2.New(ctx, log, pip2.WithFileStore("../../testdata/pip/users", false))
	entities := models.NewEntitySet(p)

	p2 := pep.New(ctx, log)
	p3 := cedar_embedded.NewController(pdp.WithLogger(log))

	authenticator := authentication2.NewBCrypt(authentication2.WithContext(ctx), authentication2.WithLogger(log), authentication2.WithEntities(entities))

	testCases := []struct {
		name              string
		opts              []Option
		wantCtx           context.Context
		wantLog           *slog.Logger
		wantPEP           pep.PEP
		wantPDP           pdp.Controller
		wantEntities      models.EntitySet
		wantAuthenticator authentication2.Authenticator
	}{
		{name: "no options"},
		{name: "context", opts: []Option{WithContext(ctx)}, wantCtx: ctx},
		{name: "logger", opts: []Option{WithLogger(log)}, wantLog: log},
		{name: "pep", opts: []Option{WithPEP(p2)}, wantPEP: p2},
		{name: "pdp", opts: []Option{WithPDP(p3)}, wantPDP: p3},
		{name: "entities", opts: []Option{WithEntities(entities)}, wantEntities: entities},
		{name: "authenticator", opts: []Option{WithAuthenticator(authenticator)}, wantAuthenticator: authenticator},
		{name: "all", opts: []Option{WithPEP(p2), WithAuthenticator(authenticator), WithLogger(log), WithPDP(p3), WithEntities(entities), WithContext(ctx)}, wantCtx: ctx, wantLog: log, wantPEP: p2, wantPDP: p3, wantEntities: entities, wantAuthenticator: authenticator},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := &auth{}
			for i := range tc.opts {
				tc.opts[i](a)
			}

			assert.Equal(t, tc.wantCtx, a.ctx)
			assert.Equal(t, tc.wantLog, a.log)
			assert.Equal(t, tc.wantPEP, a.pep)
			assert.Equal(t, tc.wantPDP, a.pdp)
			assert.Equal(t, tc.wantEntities, a.entities)
		})
	}
}
