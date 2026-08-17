package server_test

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/server"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
)

func Test_HealthEndpoints(t *testing.T) {
	cnf := &config.Config{
		PAP: config2.PAP{
			Language: "CEDAR",
		},
	}

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := server.NewExternal(cnf, lgr)
	healthService := srv.GetHealthService()
	healthFiberApp := healthService.GetFiberApp()

	testCases := map[string]struct {
		URL                string
		ExpectedStatusCode int
	}{
		"health": {
			URL:                "/healthz",
			ExpectedStatusCode: http.StatusOK,
		},
		"live": {
			URL:                "/livez",
			ExpectedStatusCode: http.StatusOK,
		},
		"ready": {
			URL:                "/readyz",
			ExpectedStatusCode: http.StatusOK,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, tc.URL, http.NoBody)
			require.Nil(t, err)

			res, err := healthFiberApp.Test(req)
			require.Nil(t, err)

			defer res.Body.Close()

			require.Equal(t, tc.ExpectedStatusCode, res.StatusCode)
		})
	}
}
