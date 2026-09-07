package authorization

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelError)
	log := slog.New(h)

	p1, err := pip.New(ctx, log, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../testdata/pip/users", false))
	require.NoError(t, err)
	p2 := pep.New(ctx, log)
	p3 := cedar_embedded.NewController(pdp.WithLogger(log))

	authenticator := authentication.NewBCrypt(authentication.WithContext(ctx), authentication.WithLogger(log), authentication.WithEntityGetter(p1.GetEntity))

	testCases := []struct {
		name              string
		opts              []Option
		wantCtx           context.Context
		wantLog           *slog.Logger
		wantPEP           *pep.PEP
		wantPDP           pdp.Controller
		wantAuthenticator authentication.Authenticator
	}{
		{name: "no options"},
		{name: "context", opts: []Option{WithContext(ctx)}, wantCtx: ctx},
		{name: "logger", opts: []Option{WithLogger(log)}, wantLog: log},
		{name: "pep", opts: []Option{WithPEP(p2)}, wantPEP: p2},
		{name: "pdp", opts: []Option{WithPDP(p3)}, wantPDP: p3},
		{name: "entities", opts: []Option{WithEntityGetter(p1.GetEntity)}},
		{name: "authenticator", opts: []Option{WithAuthenticator(authenticator)}, wantAuthenticator: authenticator},
		{name: "all", opts: []Option{WithPEP(p2), WithAuthenticator(authenticator), WithLogger(log), WithPDP(p3), WithEntityGetter(p1.GetEntity), WithContext(ctx)}, wantCtx: ctx, wantLog: log, wantPEP: p2, wantPDP: p3, wantAuthenticator: authenticator},
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
		})
	}
}
