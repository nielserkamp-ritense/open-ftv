package pdp

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
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

			p1 := pap.New(nil, logger)
			require.NotNil(t, p1)

			p2 := pip.New(pip.Config{Logger: logger})
			require.NotNil(t, p2)

			p3 := pep.New(nil, logger)
			require.NotNil(t, p3)

			b := NewBase(WithNameVersion(tc.id, tc.version), WithLogger(logger), WithPAP(p1), WithPIP(p2), WithPEP(p3))
			require.NotNil(t, b)

			assert.Equal(t, tc.want, b.String())
			assert.Equal(t, tc.id, b.Name())
			assert.Equal(t, tc.version, b.Version())
			assert.Equal(t, logger, b.Logger())
			assert.Equal(t, p1, b.PAP())
			assert.Equal(t, p2, b.PIP())
			assert.Equal(t, p3, b.PEP())
			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
		})
	}
}
