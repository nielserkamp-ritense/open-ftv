//go:build integration

package server

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
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

func Test_Put_Policy(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := NewExternal(cnf, lgr)
	mainFiberApp := srv.GetMainService().GetFiberApp()

	const policyData = "permit (\n    principal,\n    action,\n    resource\n);\n"

	t.Run("update_non-existing", func(t *testing.T) {
		policy := oas.Policy{
			Language: "cedar",
			Data:     policyData,
		}

		res := putPolicy(t, mainFiberApp, "aaaaaaaa-0000-0000-0000-000000000000", policy)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("invalid_id", func(t *testing.T) {
		policy := oas.Policy{
			Language: "cedar",
			Data:     policyData,
		}

		invalidID := strings.Repeat("x", 41)

		res := putPolicy(t, mainFiberApp, invalidID, policy)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("id_mismatch", func(t *testing.T) {
		mustCreatePolicy(t, mainFiberApp, "33333333-3333-3333-3333-333333333333", oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{Title: "Put id mismatch base"},
		})

		policy := oas.Policy{
			Id:       "99999999-9999-9999-9999-999999999999",
			Language: "cedar",
			Data:     policyData,
		}

		res := putPolicy(t, mainFiberApp, "33333333-3333-3333-3333-333333333333", policy)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("malformed_json", func(t *testing.T) {
		mustCreatePolicy(t, mainFiberApp, "44444444-4444-4444-4444-444444444444", oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{Title: "Put malformed base"},
		})

		req, err := http.NewRequest(http.MethodPut, "/v1/policy/44444444-4444-4444-4444-444444444444", bytes.NewReader([]byte("{invalid-json")))
		require.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("happy_flow", func(t *testing.T) {
		mustCreatePolicy(t, mainFiberApp, "11111111-1111-1111-1111-111111111111", oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{Title: "Update original title"},
		})

		updated := oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{
				Title:       "Update new title",
				Description: "Now with a description",
				Tags:        []string{"updated"},
			},
		}

		res := putPolicy(t, mainFiberApp, "11111111-1111-1111-1111-111111111111", updated)
		require.Equal(t, http.StatusOK, res.StatusCode)

		got := decodePolicy(t, res)

		require.Equal(t, "11111111-1111-1111-1111-111111111111", got.Id)
		require.Equal(t, "Update new title", got.Metadata.Title)
		require.Equal(t, "Now with a description", got.Metadata.Description)
		require.Equal(t, []string{"updated"}, got.Metadata.Tags)

		// ensure the changes are persisted
		fetched := decodePolicy(t, getPolicy(t, mainFiberApp, "11111111-1111-1111-1111-111111111111"))
		require.Equal(t, "Update new title", fetched.Metadata.Title)
		require.Equal(t, "Now with a description", fetched.Metadata.Description)
	})

	t.Run("happy_flow_with_force_upsert", func(t *testing.T) {
		policy := oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{Title: "Created via PUT"},
		}

		res := putPolicyWithQuery(t, mainFiberApp, "22222222-2222-2222-2222-222222222222", policy, "forceUpsert=true")
		require.Equal(t, http.StatusOK, res.StatusCode)

		got := decodePolicy(t, res)
		require.Equal(t, "22222222-2222-2222-2222-222222222222", got.Id)
		require.Equal(t, "concept", got.Status)
		require.Equal(t, "Created via PUT", got.Metadata.Title)

		fetched := getPolicy(t, mainFiberApp, "22222222-2222-2222-2222-222222222222")
		require.Equal(t, http.StatusOK, fetched.StatusCode)
	})
}

func putPolicy(t *testing.T, app *fiber.App, id string, policy oas.Policy) *http.Response {
	t.Helper()

	return putPolicyWithQuery(t, app, id, policy, "")
}

func putPolicyWithQuery(t *testing.T, app *fiber.App, id string, policy oas.Policy, query string) *http.Response {
	t.Helper()

	body, err := json.Marshal(policy)
	require.Nil(t, err)

	url := "/v1/policy/" + id
	if query != "" {
		url += "?" + query
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	require.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	require.Nil(t, err)

	return res
}
