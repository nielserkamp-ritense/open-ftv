package server

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	eam_fiber "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

// newHealthOnlyService builds only the health service.
// Health checks never touch the PAP/PIP, so this skips the main/bundle services and the persistence backend
// they'd otherwise require.
func newHealthOnlyService(cnf *config.Config, logger *slog.Logger) server.Service {
	s := &Services{cfg: cnf, logger: logger, chk: handle.NewChecks()}
	s.chk.SetHealth(true)
	s.chk.SetAlive(true)
	s.chk.SetReady(true)

	return eam_fiber.New(
		logger,
		s.initHealthRoutes,
		server.WithDefaults(),
		server.WithHostPort(cnf.HealthHost, cnf.HealthPort),
		server.WithAppName(config.AppName),
		server.WithSvcName("health"),
		server.WithTimeouts(cnf.HealthRead, cnf.HealthWrite, cnf.HealthIdle),
		server.WithMaxBody(cnf.HealthMaxBody),
		server.WithRecovery(),
		server.WithSecurity(),
		server.WithCORS(cnf.HealthOrigins, cnf.HealthHeaders),
	)
}

func Test_HealthEndpoints(t *testing.T) {
	cnf := &config.Config{}
	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	healthFiberApp := newHealthOnlyService(cnf, lgr).GetFiberApp()

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
