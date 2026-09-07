//go:build integration

package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

func Test_Get_Policies(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := NewExternal(cnf, lgr)
	mainFiberApp := srv.GetMainService().GetFiberApp()

	t.Run("no_data", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/policies", nil)
		require.Nil(t, err)

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.Nil(t, err)
		defer res.Body.Close()

		require.JSONEq(t, "[]", string(body))
	})

	t.Run("with_a_single_policy", func(t *testing.T) {
		policy := oas.Policy{
			Language: "cedar",
			Data:     "permit (\n    principal,\n    action,\n    resource\n);\n",
			Metadata: oas.Metadata{
				Title: "My first policy",
			},
		}

		body, err := json.Marshal(policy)
		require.Nil(t, err)

		postReq, err := http.NewRequest(http.MethodPost, "/v1/policy/11111111-1111-1111-1111-111111111111", bytes.NewReader(body))
		require.Nil(t, err)

		postReq.Header.Set("Content-Type", "application/json")

		postRes, err := mainFiberApp.Test(postReq)
		require.Nil(t, err)

		require.Equal(t, http.StatusCreated, postRes.StatusCode)

		req, err := http.NewRequest(http.MethodGet, "/v1/policies", nil)
		require.Nil(t, err)

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		policies := decodePolicies(t, res)
		require.Len(t, policies, 1)
		require.Equal(t, "11111111-1111-1111-1111-111111111111", policies[0].Id)
		require.Equal(t, "cedar", policies[0].Language)
		require.Equal(t, "My first policy", policies[0].Metadata.Title)
	})
}

func decodePolicies(t *testing.T, res *http.Response) []oas.Policy {
	t.Helper()

	body, err := io.ReadAll(res.Body)
	require.Nil(t, err)
	defer func() { require.NoError(t, res.Body.Close()) }()

	var policies []oas.Policy
	require.NoError(t, json.Unmarshal(body, &policies))

	return policies
}
