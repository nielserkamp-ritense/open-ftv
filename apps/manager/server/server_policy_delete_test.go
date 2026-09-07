//go:build integration

package server

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

func Test_NewServer_Delete_Policy(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := NewExternal(cnf, lgr)
	mainFiberApp := srv.GetMainService().GetFiberApp()

	const policyData = "permit (\n    principal,\n    action,\n    resource\n);\n"

	t.Run("not_found", func(t *testing.T) {
		// Deleting an unknown policy is a 404 unless ignoreMissing is set.
		res := deletePolicy(t, mainFiberApp, "22222222-2222-2222-2222-222222222222")
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("ignore_missing", func(t *testing.T) {
		const id = "33333333-3333-3333-3333-333333333333"

		// With ignoreMissing=true, deleting an unknown policy succeeds
		// idempotently and echoes back an (empty) policy carrying the id.
		res := deletePolicyWithQuery(t, mainFiberApp, id, "ignoreMissing=true")
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, id, decodePolicy(t, res).Id)
	})

	t.Run("invalid_id", func(t *testing.T) {
		res := deletePolicy(t, mainFiberApp, strings.Repeat("x", 41))
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("happy_flow", func(t *testing.T) {
		const id = "11111111-1111-1111-1111-111111111111"

		mustCreatePolicy(t, mainFiberApp, id, oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{Title: "Delete happy flow"},
		})

		res := deletePolicy(t, mainFiberApp, id)
		require.Equal(t, http.StatusOK, res.StatusCode)

		// After deletion the policy must no longer be retrievable.
		fetched := getPolicy(t, mainFiberApp, id)
		require.Equal(t, http.StatusNotFound, fetched.StatusCode)
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
