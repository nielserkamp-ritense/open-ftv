//go:build integration

package server_test

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/server"
)

func Test_NewServer_Delete_Policy(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := server.NewExternal(cnf, lgr)
	mainFiberApp := srv.GetMainService().GetFiberApp()

	t.Run("not_found", func(t *testing.T) {
		res := deletePolicy(t, mainFiberApp, "22222222-2222-2222-2222-222222222222")
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("unknown_policy", func(t *testing.T) {
		res := deletePolicyWithQuery(t, mainFiberApp, "33333333-3333-3333-3333-333333333333", "ignoreMissing=true")
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, "33333333-3333-3333-3333-333333333333", decodePolicy(t, res).Id)
	})

	t.Run("invalid_id", func(t *testing.T) {
		invalidID := strings.Repeat("x", 41)
		res := deletePolicy(t, mainFiberApp, invalidID)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func deletePolicy(t *testing.T, app *fiber.App, id string) *http.Response {
	t.Helper()

	return deletePolicyWithQuery(t, app, id, "")
}

func deletePolicyWithQuery(t *testing.T, app *fiber.App, id, query string) *http.Response {
	t.Helper()

	url := "/v1/policy/" + id
	if query != "" {
		url += "?" + query
	}

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.Nil(t, err)

	res, err := app.Test(req)
	require.Nil(t, err)

	return res
}
