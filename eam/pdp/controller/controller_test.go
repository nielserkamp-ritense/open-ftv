package controller

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewBase(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			ap := pap.New(nil, logger)
			require.NotNil(t, ap)

			ip := pip.New(nil, logger)
			require.NotNil(t, ip)

			ep := pep.New(nil, logger)
			require.NotNil(t, ep)

			b := NewBase(WithNameVersion(tc.id, tc.version), WithLogger(logger), WithPAP(ap), WithPIP(ip), WithPEP(ep))
			require.NotNil(t, b)

			assert.Equal(t, tc.want, b.String())
			assert.Equal(t, tc.id, b.Name())
			assert.Equal(t, tc.version, b.Version())
			assert.Equal(t, logger, b.Logger())
			assert.Equal(t, ap, b.PAP())
			assert.Equal(t, ip, b.PIP())
			assert.Equal(t, ep, b.PEP())
			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
		})
	}
}
