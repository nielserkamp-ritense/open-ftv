package authentication

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestBCrypt_AuthenticateApiKey(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelDebug)
	log := slog.New(h)

	p := pip2.New(ctx, log, pip2.WithFileStore("../../testdata/unittest/auth", true))
	entities := models.NewEntitySet(p)

	testCases := []struct {
		name    string
		apikey  string
		wantErr bool
	}{
		{name: "no apikey", wantErr: true},
		{name: "unknown apikey", apikey: "abcde", wantErr: true},
		{name: "all good", apikey: "abcdef", wantErr: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewBCrypt(WithContext(ctx), WithLogger(log), WithEntities(entities))
			require.NotNil(t, a)

			err := a.AuthenticateApiKey(ctx, tc.apikey)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
