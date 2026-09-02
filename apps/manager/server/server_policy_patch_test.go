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
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

func Test_Patch_PolicyStatus(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := NewExternal(cnf, lgr)
	mainFiberApp := srv.GetMainService().GetFiberApp()

	const policyData = "permit (\n    principal,\n    action,\n    resource\n);\n"

	t.Run("invalid_status", func(t *testing.T) {
		mustCreatePolicy(t, mainFiberApp, "44444444-4444-4444-4444-444444444444", oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{
				Title: "Patch invalid status value",
			},
		})

		res := patchPolicyStatus(t, mainFiberApp, "44444444-4444-4444-4444-444444444444", oas.PolicyStatus{
			Status: "an-invalid-status",
		})
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("invalid_id", func(t *testing.T) {
		invalidID := strings.Repeat("x", 41)

		res := patchPolicyStatus(t, mainFiberApp, invalidID, oas.PolicyStatus{
			Status: "accepted",
		})
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("id_mismatch", func(t *testing.T) {
		mustCreatePolicy(t, mainFiberApp, "55555555-5555-5555-5555-555555555555", oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{Title: "Patch id mismatch base"},
		})

		status := oas.PolicyStatus{
			Id: "99999999-9999-9999-9999-999999999999", Status: "accepted"}

		res := patchPolicyStatus(t, mainFiberApp, "55555555-5555-5555-5555-555555555555", status)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("malformed_json", func(t *testing.T) {
		mustCreatePolicy(t, mainFiberApp, "66666666-6666-6666-6666-666666666666", oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{
				Title: "Patch malformed base",
			},
		})

		req, err := http.NewRequest(http.MethodPatch, "/v1/policy/66666666-6666-6666-6666-666666666666/status", bytes.NewReader([]byte("{invalid-json")))
		require.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("happy_flow", func(t *testing.T) {
		created := mustCreatePolicy(t, mainFiberApp, "11111111-1111-1111-1111-111111111111", oas.Policy{
			Language: "cedar",
			Data:     policyData,
			Metadata: oas.Metadata{
				Title: "Patch concept to accepted",
			},
		})
		require.Equal(t, "concept", created.Status)

		res := patchPolicyStatus(t, mainFiberApp, "11111111-1111-1111-1111-111111111111", oas.PolicyStatus{
			Status: "accepted",
		})
		require.Equal(t, http.StatusOK, res.StatusCode)

		got := decodePolicy(t, res)
		require.Equal(t, "accepted", got.Status)
		require.NotEmpty(t, got.Audit.Updated)

		// all other fields should remain equal
		require.Equal(t, "11111111-1111-1111-1111-111111111111", got.Id)
		require.Equal(t, "cedar", got.Language)
		require.Equal(t, policyData, got.Data)
		require.Equal(t, "Patch concept to accepted", got.Metadata.Title)
		require.Equal(t, oas.Principal{Id: "*SYSTEM*", Kind: "system", Name: "*SYSTEM*"}, got.Audit.CreatedBy)
		require.Equal(t, &oas.Principal{Id: "*SYSTEM*", Kind: "system", Name: "*SYSTEM*"}, got.Audit.UpdatedBy)
		createdAt, err := time.Parse(time.RFC3339Nano, created.Audit.Created)
		require.Nil(t, err)
		require.Equal(t, createdAt.Truncate(time.Microsecond).Format(time.RFC3339Nano), got.Audit.Created)

		// ensure the change is persisted
		thePolicy := getPolicy(t, mainFiberApp, "11111111-1111-1111-1111-111111111111")
		fetched := decodePolicy(t, thePolicy)
		require.Equal(t, "accepted", fetched.Status)
	})
}

func patchPolicyStatus(t *testing.T, app *fiber.App, id string, status oas.PolicyStatus) *http.Response {
	t.Helper()

	body, err := json.Marshal(status)
	require.Nil(t, err)

	req, err := http.NewRequest(http.MethodPatch, "/v1/policy/"+id+"/status", bytes.NewReader(body))
	require.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	require.Nil(t, err)

	return res
}
