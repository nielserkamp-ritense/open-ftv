package pdp

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestOptions(t *testing.T) {
	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	p1 := pip.New(pip.Config{
		Logger:        logger,
		NewAttributes: models.NewAttributeSet,
		NewEntities:   models.NewEntitySet,
	})
	require.NotNil(t, p1)

	p2 := pap.New(nil, logger)
	require.NotNil(t, p2)

	p3 := pep.New(nil, logger)
	require.NotNil(t, p3)

	testCases := []struct {
		name        string
		options     []Option
		wantName    string
		wantVersion string
		wantFull    string
		wantLogger  *slog.Logger
		wantPEP     pep.PEP
		wantPAP     pap.PAP
		wantPIP     pip.PIP
	}{
		{
			name: "no options",
		},
		{
			name:        "name + version",
			options:     []Option{WithNameVersion("x1", "v1")},
			wantName:    "x1",
			wantVersion: "v1",
			wantFull:    "x1 v1",
		},
		{
			name:       "logger",
			options:    []Option{WithLogger(logger)},
			wantLogger: logger,
		},
		{
			name:    "pep",
			options: []Option{WithPEP(p3)},
			wantPEP: p3,
		},
		{
			name:    "pip",
			options: []Option{WithPIP(p1)},
			wantPIP: p1,
		},
		{
			name:    "pap",
			options: []Option{WithPAP(p2)},
			wantPAP: p2,
		},
		{
			name:        "all",
			options:     []Option{WithPAP(p2), WithLogger(logger), WithPIP(p1), WithNameVersion("x1", "v1"), WithPEP(p3)},
			wantName:    "x1",
			wantVersion: "v1",
			wantFull:    "x1 v1",
			wantLogger:  logger,
			wantPEP:     p3,
			wantPAP:     p2,
			wantPIP:     p1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewBase(tc.options...)
			require.NotNil(t, got)

			assert.Equal(t, tc.wantName, got.Name())
			assert.Equal(t, tc.wantVersion, got.Version())
			assert.Equal(t, tc.wantFull, got.String())
			assert.Equal(t, tc.wantLogger, got.Logger())
			assert.Equal(t, tc.wantPEP, got.PEP())
			assert.Equal(t, tc.wantPIP, got.PIP())
			assert.Equal(t, tc.wantPAP, got.PAP())
		})
	}
}
