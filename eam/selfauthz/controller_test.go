package selfauthz

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	cerbos_api "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestBuildController_Languages(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		language string
		store    string
		wantErr  bool
	}{
		{
			name:     "unsupported language",
			language: "ai-magic",
			wantErr:  true,
		},
		{
			name:     "Cedar",
			language: "CEDAR",
			store:    "../../testdata/unittest/cedar",
		},
		{
			name:     "Cerbos",
			language: "Cerbos",
			store:    "../../testdata/unittest/cerbos",
		},
		{
			name:     "OpenFGA",
			language: "OpenFGA",
			store:    "../../testdata/unittest/openfga",
		},
		{
			name:     "OPA/Rego",
			language: "opa",
			store:    "../../testdata/unittest/opa",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
			language := models.LanguageFromString(tc.language)

			ip, err := pip.NewSelfAuthz(context.Background(), logger, "../../testdata/unittest/pip2", false, "")
			require.NoError(t, err)

			controller, err := BuildController(context.Background(), logger, language, ip, &cerbos_api.Config{}).
				WithBundledPAP(tc.language, tc.store, false).
				Build()

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, controller)

				return
			}

			require.NoError(t, err)
			assert.NotNil(t, controller)
			assert.NotNil(t, controller.GetPAP())
		})
	}
}

func TestControllerBuilder_Build_NoPAPConfigured(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))

	controller, err := BuildController(context.Background(), logger, models.CEDAR, newPIP(t, logger), &cerbos_api.Config{}).Build()

	require.NoError(t, err)
	assert.NotNil(t, controller)
	assert.Nil(t, controller.GetPAP())
}

func TestControllerBuilder_WithPAP_OverridesBundled(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	live, err := pap.NewSelfAuthz(context.Background(), logger, "cedar", "", false)
	require.NoError(t, err)

	controller, err := BuildController(context.Background(), logger, models.CEDAR, newPIP(t, logger), &cerbos_api.Config{}).
		WithBundledPAP("cedar", "../../testdata/unittest/cedar", false).
		WithPAP(live).
		Build()

	require.NoError(t, err)
	assert.Same(t, live, controller.GetPAP())
}

func TestControllerBuilder_WithBundledPAP_OverridesWithPAP(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	live, err := pap.NewSelfAuthz(context.Background(), logger, "cedar", "", false)
	require.NoError(t, err)

	controller, err := BuildController(context.Background(), logger, models.CEDAR, newPIP(t, logger), &cerbos_api.Config{}).
		WithPAP(live).
		WithBundledPAP("cedar", "../../testdata/unittest/cedar", false).
		Build()

	require.NoError(t, err)
	assert.NotSame(t, live, controller.GetPAP())
	assert.NotNil(t, controller.GetPAP())
}

func TestControllerBuilder_WithOptions_OverridesDefault(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	customPEP := pep.New(context.Background(), logger)

	controller, err := BuildController(context.Background(), logger, models.CEDAR, newPIP(t, logger), &cerbos_api.Config{}).
		WithBundledPAP("cedar", "../../testdata/unittest/cedar", false).
		WithOptions(pdp.WithPEP(customPEP)).
		Build()

	require.NoError(t, err)
	assert.Same(t, customPEP, controller.GetPEP())
}

func newPIP(t *testing.T, logger *slog.Logger) *pip.PIP {
	t.Helper()

	ip, err := pip.NewSelfAuthz(context.Background(), logger, "", false, "")
	require.NoError(t, err)

	return ip
}
