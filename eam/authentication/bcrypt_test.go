package authentication

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNewBCrypt(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	p, err := pip2.New(ctx, log, pip2.WithKeyValueDB(memory.New(), ""), pip2.WithFileStore("../../testdata/pip/users", false))
	require.NoError(t, err)

	testCases := []struct {
		name    string
		opts    []Option
		wantCtx context.Context
		wantLog *slog.Logger
	}{
		{name: "no options"},
		{name: "context", opts: []Option{WithContext(ctx)}, wantCtx: ctx},
		{name: "logger", opts: []Option{WithLogger(log)}, wantLog: log},
		{name: "entities", opts: []Option{WithEntityGetter(p.GetEntity)}},
		{name: "all", opts: []Option{WithLogger(log), WithEntityGetter(p.GetEntity), WithContext(ctx)}, wantCtx: ctx, wantLog: log},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewBCrypt(tc.opts...)
			require.NotNil(t, got)

			got2, ok := got.(*bcryptAuth)
			require.True(t, ok)
			require.NotNil(t, got2)

			assert.NotNil(t, got2.ctx)
			assert.NotNil(t, got2.log)
			assert.NotNil(t, got2.getEntity)

			if tc.wantCtx != nil {
				assert.Equal(t, tc.wantCtx, got2.ctx)
			}
			if tc.wantLog != nil {
				assert.Equal(t, tc.wantLog, got2.log)
			}
		})
	}
}

func TestBCrypt_AuthenticateUser(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelDebug)
	log := slog.New(h)

	p, err := pip2.New(ctx, log, pip2.WithKeyValueDB(memory.New(), ""), pip2.WithFileStore("../../testdata/unittest/auth", true))
	require.NoError(t, err)

	testCases := []struct {
		name    string
		user    string
		pswd    string
		wantErr bool
	}{
		{name: "no user", pswd: "mouse", wantErr: true},
		{name: "no pswd", user: "mickey", wantErr: true},
		{name: "unknown user", user: "minnie", pswd: "mouse", wantErr: true},
		{name: "bad pswd", user: "mickey", pswd: "mousse", wantErr: true},
		{name: "bad entity", user: "donald", pswd: "duck", wantErr: true},
		{name: "all good", user: "mickey", pswd: "mouse", wantErr: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := NewBCrypt(WithContext(ctx), WithLogger(log), WithEntityGetter(p.GetEntity))
			require.NotNil(t, b)

			err := b.AuthenticateUser(ctx, tc.user, tc.pswd)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
