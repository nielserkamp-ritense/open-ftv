package config

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestBuildSelfAuthzController_Languages(t *testing.T) {
	t.Parallel()

	pipCfg := &PIP{Store: "../../testdata/unittest/pip2"}
	cerbosCfg := &Cerbos{}

	testCases := []struct {
		name    string
		papCfg  *PAP
		wantErr bool
	}{
		{
			name:    "unsupported language",
			papCfg:  &PAP{Language: "ai-magic"},
			wantErr: true,
		},
		{
			name:   "Cedar",
			papCfg: &PAP{Language: "CEDAR", Store: "../../testdata/unittest/cedar"},
		},
		{
			name:   "Cerbos",
			papCfg: &PAP{Language: "Cerbos", Store: "../../testdata/unittest/cerbos"},
		},
		{
			name:   "OpenFGA",
			papCfg: &PAP{Language: "OpenFGA", Store: "../../testdata/unittest/openfga"},
		},
		{
			name:   "OPA/Rego",
			papCfg: &PAP{Language: "opa", Store: "../../testdata/unittest/opa"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
			language := models.LanguageFromString(tc.papCfg.Language)

			controller, ip, err := BuildSelfAuthzController(context.Background(), logger, language, pipCfg, cerbosCfg).
				WithBundledPAP(tc.papCfg).
				Build()

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, controller)

				return
			}

			require.NoError(t, err)
			assert.NotNil(t, controller)
			assert.NotNil(t, ip)
			assert.NotNil(t, controller.GetPAP())
		})
	}
}

func TestSelfAuthzControllerBuilder_Build_NoPAPConfigured(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))

	controller, ip, err := BuildSelfAuthzController(context.Background(), logger, models.CEDAR, &PIP{}, &Cerbos{}).Build()

	require.NoError(t, err)
	assert.NotNil(t, controller)
	assert.NotNil(t, ip)
	assert.Nil(t, controller.GetPAP())
}

func TestSelfAuthzControllerBuilder_WithPAP_OverridesBundled(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	live, err := (&PAP{Language: "cedar"}).NewSelfAuthzPAP(context.Background(), logger)
	require.NoError(t, err)

	controller, _, err := BuildSelfAuthzController(context.Background(), logger, models.CEDAR, &PIP{}, &Cerbos{}).
		WithBundledPAP(&PAP{Language: "cedar", Store: "../../testdata/unittest/cedar"}).
		WithPAP(live).
		Build()

	require.NoError(t, err)
	assert.Same(t, live, controller.GetPAP())
}

func TestSelfAuthzControllerBuilder_WithBundledPAP_OverridesWithPAP(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	live, err := (&PAP{Language: "cedar"}).NewSelfAuthzPAP(context.Background(), logger)
	require.NoError(t, err)

	controller, _, err := BuildSelfAuthzController(context.Background(), logger, models.CEDAR, &PIP{}, &Cerbos{}).
		WithPAP(live).
		WithBundledPAP(&PAP{Language: "cedar", Store: "../../testdata/unittest/cedar"}).
		Build()

	require.NoError(t, err)
	assert.NotSame(t, live, controller.GetPAP())
	assert.NotNil(t, controller.GetPAP())
}

func TestSelfAuthzControllerBuilder_WithOptions_OverridesDefault(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	customPEP := pep.New(context.Background(), logger)

	controller, _, err := BuildSelfAuthzController(context.Background(), logger, models.CEDAR, &PIP{}, &Cerbos{}).
		WithBundledPAP(&PAP{Language: "cedar", Store: "../../testdata/unittest/cedar"}).
		WithOptions(pdp.WithPEP(customPEP)).
		Build()

	require.NoError(t, err)
	assert.Same(t, customPEP, controller.GetPEP())
}
