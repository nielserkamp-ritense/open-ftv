//go:build integration

package server_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/server"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

func Test_Post_Policy(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := server.NewExternal(cnf, lgr)
	mainFiberApp := srv.GetMainService().GetFiberApp()

	const policyData = "permit (\n    principal,\n    action,\n    resource\n);\n"

	t.Run("conflict_on_existing", func(t *testing.T) {
		policy := oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{Title: "Conflict policy"},
		}

		res := postPolicy(t, mainFiberApp, "22222222-2222-2222-2222-222222222222", policy)
		require.Equal(t, http.StatusCreated, res.StatusCode)

		res = postPolicy(t, mainFiberApp, "22222222-2222-2222-2222-222222222222", policy)
		require.Equal(t, http.StatusConflict, res.StatusCode)
	})

	t.Run("force_upsert_updates_existing", func(t *testing.T) {
		policy := oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{
				Title: "Original title",
			},
		}

		res := postPolicy(t, mainFiberApp, "33333333-3333-3333-3333-333333333333", policy)
		require.Equal(t, http.StatusCreated, res.StatusCode)

		policy.Metadata.Title = "Updated title"

		res = postPolicyWithQuery(t, mainFiberApp, "33333333-3333-3333-3333-333333333333", policy, "forceUpsert=true")
		require.Equal(t, http.StatusCreated, res.StatusCode)

		got := decodePolicy(t, res)
		require.Equal(t, "Updated title", got.Metadata.Title)

		// ensure the update is persisted
		fetched := getPolicy(t, mainFiberApp, "33333333-3333-3333-3333-333333333333")
		require.Equal(t, http.StatusOK, fetched.StatusCode)
		require.Equal(t, "Updated title", decodePolicy(t, fetched).Metadata.Title)
	})

	t.Run("invalid_id", func(t *testing.T) {
		policy := oas.Policy{
			Language: "cedar",
			Data:     policyData,
		}

		invalidID := strings.Repeat("x", 41)

		res := postPolicy(t, mainFiberApp, invalidID, policy)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("id_mismatch", func(t *testing.T) {
		policy := oas.Policy{
			Id:       "99999999-9999-9999-9999-999999999999",
			Language: "cedar",
			Data:     policyData,
		}

		res := postPolicy(t, mainFiberApp, "44444444-4444-4444-4444-444444444444", policy)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("missing_language", func(t *testing.T) {
		policy := oas.Policy{
			Data: policyData,
		}

		res := postPolicy(t, mainFiberApp, "55555555-5555-5555-5555-555555555555", policy)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("missing_data_and_id", func(t *testing.T) {
		policy := oas.Policy{
			Language: "cedar",
		}

		res := postPolicy(t, mainFiberApp, "66666666-6666-6666-6666-666666666666", policy)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("malformed_json", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/v1/policy/77777777-7777-7777-7777-777777777777", bytes.NewReader([]byte("{invalid-json")))
		require.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("happy_flow", func(t *testing.T) {
		policy := oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{
				Title:       "My first policy",
				Description: "The very first policy",
				RvvaId:      "rvva-1",
				Tags:        []string{"first"},
			},
		}

		res := postPolicy(t, mainFiberApp, "11111111-1111-1111-1111-111111111111", policy)
		require.Equal(t, http.StatusCreated, res.StatusCode)

		got := decodePolicy(t, res)
		require.Equal(t, "11111111-1111-1111-1111-111111111111", got.Id)
		require.Equal(t, "cedar", got.Language)
		require.Equal(t, "concept", got.Status)
		require.Equal(t, policyData, got.Data)
		require.Equal(t, "My first policy", got.Metadata.Title)
		require.Equal(t, "The very first policy", got.Metadata.Description)
		require.Equal(t, "rvva-1", got.Metadata.RvvaId)
		require.Equal(t, []string{"first"}, got.Metadata.Tags)
		require.Equal(t, "*SYSTEM*", got.Audit.CreatedBy)
		require.NotEmpty(t, got.Audit.Created)
	})
}

func postPolicy(t *testing.T, app *fiber.App, id string, policy oas.Policy) *http.Response {
	t.Helper()

	return postPolicyWithQuery(t, app, id, policy, "")
}

func postPolicyWithQuery(t *testing.T, app *fiber.App, id string, policy oas.Policy, query string) *http.Response {
	t.Helper()

	body, err := json.Marshal(policy)
	require.Nil(t, err)

	url := "/v1/policy/" + id
	if query != "" {
		url += "?" + query
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	require.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	require.Nil(t, err)

	return res
}

func mustCreatePolicy(t *testing.T, app *fiber.App, id string, policy oas.Policy) oas.Policy {
	t.Helper()

	res := postPolicy(t, app, id, policy)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	return decodePolicy(t, res)
}

func getPolicy(t *testing.T, app *fiber.App, id string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "/v1/policy/"+id, nil)
	require.Nil(t, err)

	res, err := app.Test(req)
	require.Nil(t, err)

	return res
}
