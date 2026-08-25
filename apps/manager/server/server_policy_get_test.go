//go:build integration

package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/server"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

func Test_Get_Policy(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := server.NewExternal(cnf, lgr)
	mainFiberApp := srv.GetMainService().GetFiberApp()

	t.Run("not_found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/policy/22222222-2222-2222-2222-222222222222", nil)
		require.Nil(t, err)

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("invalid_id", func(t *testing.T) {
		invalidID := strings.Repeat("x", 41)
		req, err := http.NewRequest(http.MethodGet, "/v1/policy/"+invalidID, nil)
		require.Nil(t, err)

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("happy_flow", func(t *testing.T) {
		const policyData = "permit (\n    principal,\n    action,\n    resource\n);\n"

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

		body, err := json.Marshal(policy)
		require.Nil(t, err)

		postReq, err := http.NewRequest(http.MethodPost, "/v1/policy/11111111-1111-1111-1111-111111111111", bytes.NewReader(body))
		require.Nil(t, err)

		postReq.Header.Set("Content-Type", "application/json")

		postRes, err := mainFiberApp.Test(postReq)
		require.Nil(t, err)
		require.Equal(t, http.StatusCreated, postRes.StatusCode)

		req, err := http.NewRequest(http.MethodGet, "/v1/policy/11111111-1111-1111-1111-111111111111", nil)
		require.Nil(t, err)

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		got := decodePolicy(t, res)
		require.Equal(t, "11111111-1111-1111-1111-111111111111", got.Id)
		require.Equal(t, "cedar", got.Language)
		require.Equal(t, "concept", got.Status)
		require.Equal(t, policyData, got.Data)
		require.Equal(t, "My first policy", got.Metadata.Title)
		require.Equal(t, "The very first policy", got.Metadata.Description)
		require.Equal(t, "rvva-1", got.Metadata.RvvaId)
		require.Equal(t, []string{"first"}, got.Metadata.Tags)
		require.Equal(t, oas.Principal{Id: "*SYSTEM*", Kind: "system", Name: "*SYSTEM*"}, got.Audit.CreatedBy)
		require.Equal(t, &oas.Principal{Id: "*SYSTEM*", Kind: "system", Name: "*SYSTEM*"}, got.Audit.UpdatedBy)
		require.NotEmpty(t, got.Audit.Created)
		require.NotEmpty(t, got.Audit.Updated)
	})
}

func decodePolicy(t *testing.T, res *http.Response) oas.Policy {
	t.Helper()

	body, err := io.ReadAll(res.Body)
	require.Nil(t, err)
	defer func() { require.NoError(t, res.Body.Close()) }()

	var policy oas.Policy
	require.NoError(t, json.Unmarshal(body, &policy))

	return policy
}
