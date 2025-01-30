package pdp

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewBase(t *testing.T) {
	testCases := []struct {
		name    string
		id      string
		version string
		want    string
		wantLog int
	}{
		{
			name:    "cedar v1",
			id:      "cedar",
			version: "v1",
			want:    "cedar v1",
			wantLog: 2,
		},
		{
			name:    "OPA v0.7",
			id:      "OPA",
			version: "v0.7",
			want:    "OPA v0.7",
			wantLog: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			b := NewBase(WithNameVersion(tc.id, tc.version), WithLogger(logger))
			require.NotNil(t, b)

			assert.Equal(t, tc.want, b.String())
			assert.Equal(t, tc.id, b.Name())
			assert.Equal(t, tc.version, b.Version())
			assert.Equal(t, logger, b.Logger())
			assert.Nil(t, b.pap)
			assert.Nil(t, b.pip)
			assert.Nil(t, b.logboek)

			p1 := pap.New(nil, logger, nil)
			require.NotNil(t, p1)

			b.SetPAP(p1)
			assert.Equal(t, p1, b.PAP())

			p2 := pip.New(pip.Config{Logger: logger})
			require.NotNil(t, p2)

			b.SetPIP(p2)
			assert.Equal(t, p2, b.PIP())

			assert.Nil(t, b.Logboek())

			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
		})
	}
}
