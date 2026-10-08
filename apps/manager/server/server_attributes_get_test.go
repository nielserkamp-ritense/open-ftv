//go:build integration

package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

func Test_Get_Attributes(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := NewExternal(cnf, lgr)
	mainFiberApp := srv.GetMainService().GetFiberApp()

	t.Run("no_data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/attributes", http.NoBody)

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)

		defer res.Body.Close()

		require.Equal(t, http.StatusOK, res.StatusCode)

		attrs := decodeAttributes(t, res)
		require.Empty(t, attrs)
	})

	t.Run("with_a_single_attribute", func(t *testing.T) {
		attribute := oas.Attribute{
			Key:   "my-first-key",
			Value: 42,
			Type:  "integer",
		}

		body, err := json.Marshal(attribute)
		require.Nil(t, err)

		postReq := httptest.NewRequest(http.MethodPost, "/v1/attribute/my-first-key", bytes.NewReader(body))

		postReq.Header.Set("Content-Type", "application/json")

		postRes, err := mainFiberApp.Test(postReq)
		require.Nil(t, err)

		defer postRes.Body.Close()

		require.Equal(t, http.StatusCreated, postRes.StatusCode)

		req := httptest.NewRequest(http.MethodGet, "/v1/attributes", http.NoBody)

		res, err := mainFiberApp.Test(req)
		require.Nil(t, err)

		defer res.Body.Close()

		require.Equal(t, http.StatusOK, res.StatusCode)

		attrs := decodeAttributes(t, res)
		require.Len(t, attrs, 1)
		require.Equal(t, "my-first-key", attrs[0].Key)
		require.Equal(t, float64(42), attrs[0].Value)
	})
}

func decodeAttributes(t *testing.T, res *http.Response) []oas.Attribute {
	t.Helper()

	body, err := io.ReadAll(res.Body)
	require.Nil(t, err)

	var attrs []oas.Attribute
	require.NoError(t, json.Unmarshal(body, &attrs))

	return attrs
}
