//go:build integration

package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
)

// Test_ADL_SearchByTraceAndID covers the two searches the Logboek links rely on: the decisions of a trace ID,
// however old, and a single decision by its id.
func Test_ADL_SearchByTraceAndID(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	cnf.PAP.Store = "../../../testdata/apps/manager/policies/cedar"
	cnf.PIP.Store = "../../../testdata/apps/manager/data/entities"

	ip := newIDP(t)
	cnf.OIDC = config2.OIDC{JWKSURL: ip.srv.URL, Issuer: "https://idp.test", Audience: "openftv"}

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))
	app := NewExternal(cnf, lgr).GetMainService().GetFiberApp()
	auditor := ip.token(t, "auditor-user", "auditor")

	const (
		trace = "4bf92f3577b34da6a3ce929d0e0e4736"
		other = "0af7651916cd43dd8448eb211c80319c"
	)

	now := time.Now()
	recent := insertDecision(t, cnf.ADLDatabaseURL, now.Add(-time.Hour), trace, "00f067aa0ba902b7")
	// Older than the 7 days a search without a trace ID may span.
	old := insertDecision(t, cnf.ADLDatabaseURL, now.AddDate(0, 0, -30), trace, "b7ad6b7169203331")
	insertDecision(t, cnf.ADLDatabaseURL, now.Add(-time.Hour), other, "53995c3f42cd8ad8")

	search := func(query string) []int64 {
		res := call(t, app, http.MethodGet, "/v1/adl/entries?"+query, auditor, nil)
		defer res.Body.Close()

		require.Equal(t, http.StatusOK, res.StatusCode, query)

		var entries []struct {
			ID      int64  `json:"id"`
			TraceID string `json:"traceId"`
		}
		require.NoError(t, json.NewDecoder(res.Body).Decode(&entries))

		ids := make([]int64, 0, len(entries))
		for _, e := range entries {
			require.Equal(t, trace, e.TraceID, query)

			ids = append(ids, e.ID)
		}

		return ids
	}

	t.Run("a trace ID finds all its decisions, however old, in any case", func(t *testing.T) {
		require.ElementsMatch(t, []int64{recent, old}, search("traceId="+strings.ToUpper(trace)))
	})

	t.Run("an id finds its decision, however old", func(t *testing.T) {
		require.Equal(t, []int64{old}, search("id="+strconv.FormatInt(old, 10)))
	})
}

func insertDecision(t *testing.T, dsn string, at time.Time, traceID, spanID string) int64 {
	t.Helper()

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)

	defer func() { _ = conn.Close(ctx) }()

	var id int64

	err = conn.QueryRow(ctx,
		`INSERT INTO decision (timestamp, trace_id, span_id, event_name, status, request_type, body)
		 VALUES ($1, $2, $3, 'adl.access_evaluation', 'Ok', 1, '{"adl.core.response": {"decision": true}}')
		 RETURNING id`,
		at.UnixMilli(), traceID, spanID,
	).Scan(&id)
	require.NoError(t, err)

	return id
}
