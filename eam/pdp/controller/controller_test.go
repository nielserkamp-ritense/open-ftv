package controller

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
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

			ap, err := pap.New(t.Context(), logger, pap.WithKeyValueDB(memory.New(), ""))
			require.NoError(t, err)
			require.NotNil(t, ap)

			ip, err := pip.New(t.Context(), logger, pip.WithKeyValueDB(memory.New(), ""))
			require.NoError(t, err)
			require.NotNil(t, ip)

			ep := pep.New(nil, logger)
			require.NotNil(t, ep)

			b := NewBase(WithNameVersion(tc.id, tc.version), WithLogger(logger), WithPAP(ap), WithPIP(ip), WithPEP(ep))
			require.NotNil(t, b)

			assert.Equal(t, tc.want, b.String())
			assert.Equal(t, tc.id, b.Name)
			assert.Equal(t, tc.version, b.Version)
			assert.Equal(t, logger, b.GetLogger())
			assert.Equal(t, ap, b.GetPAP())
			assert.Equal(t, ip, b.GetPIP())
			assert.Equal(t, ep, b.GetPEP())
			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
		})
	}
}

func TestBase_Map(t *testing.T) {
	t.Parallel()

	t.Run("base map", func(t *testing.T) {
		f := func(parc *models.PARC, _ ...mapping.Option) *models.PARC {
			parc.Principal = models.NewEntity("hello", "world", nil)
			return parc
		}

		c := &Base{mappers: []mapping.Mapper{f}}

		parc := c.Map(&models.PARC{})
		require.NotNil(t, parc)
		assert.Equal(t, "hello", parc.Principal.Type())
		assert.Equal(t, "world", parc.Principal.ID())
	})
}
