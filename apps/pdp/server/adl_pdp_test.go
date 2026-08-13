//go:build integration

package server

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Verifies the PDP's share of the ADL Level 1 claims in docs/adr/0005-logius-adl-level1-compliancy.md —
// record content and receiving-side trace context, not the gateway's propagation half.

const defaultADLTestPostgresURL = "postgres://user:password@localhost:5432/openftv?sslmode=disable"

const goodEvaluationItem = `"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_update","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}`

const goodSubject = `{` + goodEvaluationItem + `}`

// newADLTestConfig returns a PDP config wired to a fresh, isolated Postgres-backed ADL for one test, and
// the DSN of that database for direct verification queries.
func newADLTestConfig(t *testing.T) (*config.Config, string) {
	t.Helper()

	baseURL := os.Getenv("MANAGER_TEST_POSTGRES_URL")
	if baseURL == "" {
		baseURL = defaultADLTestPostgresURL
	}

	ctx := context.Background()

	admin, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Skipf("postgres not reachable at %s (start it with `docker compose -f docker/postgres.yaml up -d`): %v", baseURL, err)
	}
	t.Cleanup(func() { _ = admin.Close(ctx) })

	dbName := adlTestDatabaseName(t, t.Name())

	if _, err = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName+" WITH (FORCE)"); err != nil {
		t.Fatalf("failed to drop existing test database %q: %v", dbName, err)
	}
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("failed to create test database %q: %v", dbName, err)
	}

	dsn := withADLDatabase(baseURL, dbName)

	cfg := &config.Config{
		PAP: config2.PAP{Language: "cedar", Store: "../../../testdata/unittest/cedar"},
		PIP: config2.PIP{Store: "../../../testdata/pip", StoreRecurse: true},
		DecisionLog: config2.DecisionLog{
			Type:          "postgresql",
			Service:       "adl-compliance-test",
			Timeout:       2 * time.Second,
			PgURL:         dsn,
			PgMaxLife:     time.Minute,
			PgMaxConn:     5,
			MigrateSource: "*embed*",
			MigrateAuto:   true,
		},
	}

	return cfg, dsn
}

var nonDatabaseNameChar = regexp.MustCompile(`[^a-z0-9_]+`)

// adlTestDatabaseName fits a (possibly long) test name into Postgres's 63-char identifier limit, hashed so
// truncated names can't collide.
func adlTestDatabaseName(t *testing.T, name string) string {
	t.Helper()

	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	suffix := fmt.Sprintf("_%08x", h.Sum32())

	safe := nonDatabaseNameChar.ReplaceAllString(strings.ToLower(name), "_")

	const maxLengthDatabaseName = 63 // Postgres identifier length limit.
	const prefix = "adl_"

	if maxPrefixLen := maxLengthDatabaseName - len(prefix) - len(suffix); len(safe) > maxPrefixLen {
		safe = safe[:maxPrefixLen]
	}

	return prefix + safe + suffix
}

func withADLDatabase(dsn, name string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}

	u.Path = "/" + name
	return u.String()
}

// newADLTestApp builds the AuthZEN fiber routes wired to a Postgres-backed ADL, using the real PDP's own
// construction path rather than a hand-rolled substitute.
func newADLTestApp(t *testing.T, cfg *config.Config) *fiber.App {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := &Services{ctx: context.Background(), cfg: cfg, logger: logger, l: models.LanguageFromString(cfg.Language)}

	auth := s.newAuth("")
	require.NotNil(t, auth, "failed to build auth handler (see logger output)")

	app := fiber.New()
	app.Post("/authzen/v1/evaluation", auth.zen.Evaluation)
	app.Post("/authzen/v1/evaluations", auth.zen.Evaluations)
	app.Post("/authzen/v1/search/subject", auth.zen.SearchSubject)
	app.Post("/authzen/v1/search/action", auth.zen.SearchAction)
	app.Post("/authzen/v1/search/resource", auth.zen.SearchResource)

	return app
}

// postEvaluation posts the known-allowed AuthZEN fixture and returns the raw HTTP response.
func postEvaluation(t *testing.T, app *fiber.App, headers map[string]string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/authzen/v1/evaluation", bytes.NewReader([]byte(goodSubject)))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	return resp
}

// decisionRow is the subset of the `decision` table columns these tests inspect directly.
type decisionRow struct {
	TraceID      string
	SpanID       string
	ParentSpanID *string
	EventName    string
	Status       string
	Timestamp    int64
	Body         []byte
	Attributes   []byte
	Resource     []byte
}

// queryDecisionByTraceID returns the most recent decision row for a trace, failing the test if none exists.
func queryDecisionByTraceID(t *testing.T, dsn, traceID string) decisionRow {
	t.Helper()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer conn.Close(ctx)

	var row decisionRow
	err = conn.QueryRow(ctx,
		`SELECT trace_id, span_id, parent_span_id, event_name, status, timestamp, body, attributes, resource
		 FROM decision WHERE trace_id = $1 ORDER BY id DESC LIMIT 1`,
		traceID,
	).Scan(&row.TraceID, &row.SpanID, &row.ParentSpanID, &row.EventName, &row.Status, &row.Timestamp, &row.Body, &row.Attributes, &row.Resource)
	require.NoError(t, err, "no decision row found for trace_id=%s", traceID)

	return row
}

func countDecisionsByTraceID(t *testing.T, dsn, traceID string) int {
	t.Helper()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer conn.Close(ctx)

	var count int
	err = conn.QueryRow(ctx, "SELECT count(*) FROM decision WHERE trace_id = $1", traceID).Scan(&count)
	require.NoError(t, err)

	return count
}

// TestADLCompliance_RecordFields covers ADR "The record" (§3.3.1-§3.3.9, §4.1.1) for a plain evaluation.
func TestADLCompliance_RecordFields(t *testing.T) {
	cfg, dsn := newADLTestConfig(t)
	app := newADLTestApp(t, cfg)

	const incomingTraceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	const incomingSpanID = "00f067aa0ba902b7"

	resp := postEvaluation(t, app, map[string]string{
		"traceparent": "00-" + incomingTraceID + "-" + incomingSpanID + "-01",
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	row := queryDecisionByTraceID(t, dsn, incomingTraceID)

	assert.Equal(t, incomingTraceID, row.TraceID, "§3.3.1: trace_id MUST be preserved")
	assert.Len(t, row.SpanID, 16, "§3.3.2: span_id MUST be a valid 16-hex span id")
	assert.NotEqual(t, incomingSpanID, row.SpanID, "§3.2.2: the PDP MUST mint its own child span on receipt")

	require.NotNil(t, row.ParentSpanID, "§3.3.3: parent_span_id MUST be set from the incoming parent")
	assert.Equal(t, incomingSpanID, *row.ParentSpanID)

	assert.Equal(t, "adl.access_evaluation", row.EventName, "§3.3.4")
	assert.Positive(t, row.Timestamp, "§3.3.5: timestamp MUST be set")
	assert.Equal(t, "Ok", row.Status, "§3.3.6")

	assert.Contains(t, string(row.Body), `"adl.core.request"`, "§3.3.7.1, §3.3.8: request MUST be carried in body")
	assert.Contains(t, string(row.Body), `"adl.core.response"`, "§3.3.7.2, §3.3.8: response MUST be carried in body")
	assert.JSONEq(t, "{}", string(row.Attributes), "§3.3.7, §4.1.1: source references stay empty at Level 1")

	assert.Contains(t, string(row.Resource), "adl-compliance-test", "§3.3.9: producer identity MUST be set")
}

// TestADLCompliance_FSCTransactionID covers ADR "FSC TransactionID" (§3.3.7.6).
func TestADLCompliance_FSCTransactionID(t *testing.T) {
	cfg, dsn := newADLTestConfig(t)
	app := newADLTestApp(t, cfg)

	const traceID = "5bf92f3577b34da6a3ce929d0e0e4737"
	const fscTxnID = "fsc-txn-abc123"

	resp := postEvaluation(t, app, map[string]string{
		"traceparent":        "00-" + traceID + "-00f067aa0ba902b7-01",
		"fsc-transaction-id": fscTxnID,
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	row := queryDecisionByTraceID(t, dsn, traceID)
	assert.JSONEq(t, `{"adl.fsc.transaction_id":"`+fscTxnID+`"}`, string(row.Attributes))
}

// TestADLCompliance_TraceContextOnReceipt covers "Honouring the caller's trace" (§3.2.2, §3.3.3).
func TestADLCompliance_TraceContextOnReceipt(t *testing.T) {
	t.Run("no traceparent starts a root trace", func(t *testing.T) {
		cfg, dsn := newADLTestConfig(t)
		app := newADLTestApp(t, cfg)

		resp := postEvaluation(t, app, nil)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		row := queryFirstDecision(t, dsn)
		assert.Len(t, row.TraceID, 32)
		assert.Len(t, row.SpanID, 16)
		assert.Nil(t, row.ParentSpanID, "§3.3.3: parent_span_id MAY be omitted only when the record is the root of its trace")
	})

	t.Run("malformed traceparent is treated as absent, not fatal", func(t *testing.T) {
		cfg, dsn := newADLTestConfig(t)
		app := newADLTestApp(t, cfg)

		resp := postEvaluation(t, app, map[string]string{"traceparent": "not-a-traceparent"})
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		row := queryFirstDecision(t, dsn)
		assert.Len(t, row.TraceID, 32)
		assert.Nil(t, row.ParentSpanID)
	})

	t.Run("unsampled trace is still logged", func(t *testing.T) {
		cfg, dsn := newADLTestConfig(t)
		app := newADLTestApp(t, cfg)

		const traceID = "6bf92f3577b34da6a3ce929d0e0e4738"

		resp := postEvaluation(t, app, map[string]string{
			"traceparent": "00-" + traceID + "-00f067aa0ba902b7-00", // sampled=00
		})
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		row := queryDecisionByTraceID(t, dsn, traceID)
		assert.Equal(t, "Ok", row.Status, "§3.2.2: logging MUST NOT depend on the sampled bit")
	})
}

func queryFirstDecision(t *testing.T, dsn string) decisionRow {
	t.Helper()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer conn.Close(ctx)

	var row decisionRow
	err = conn.QueryRow(ctx,
		`SELECT trace_id, span_id, parent_span_id, event_name, status, timestamp, body, attributes, resource
		 FROM decision ORDER BY id DESC LIMIT 1`,
	).Scan(&row.TraceID, &row.SpanID, &row.ParentSpanID, &row.EventName, &row.Status, &row.Timestamp, &row.Body, &row.Attributes, &row.Resource)
	require.NoError(t, err, "no decision row found")

	return row
}

// TestADLCompliance_ExactlyOneRecordPerDecision covers ADR "Exactly one record per decision" (§3.2.3): a
// single Access Evaluations call, however many sub-evaluations it batches, MUST produce exactly one record.
func TestADLCompliance_ExactlyOneRecordPerDecision(t *testing.T) {
	cfg, dsn := newADLTestConfig(t)
	app := newADLTestApp(t, cfg)

	const traceID = "7bf92f3577b34da6a3ce929d0e0e4739"

	body := `{"evaluations":[{` + goodEvaluationItem + `},{` + goodEvaluationItem + `},{` + goodEvaluationItem + `}]}`

	req := httptest.NewRequest(http.MethodPost, "/authzen/v1/evaluations", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("traceparent", "00-"+traceID+"-00f067aa0ba902b7-01")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, 1, countDecisionsByTraceID(t, dsn, traceID))
}

// TestADLCompliance_RecordQueryableWithoutPolling relates to §2.3.2, but can't prove write-before-response
// ordering: this in-process harness has no real network flush to race against.
func TestADLCompliance_RecordQueryableWithoutPolling(t *testing.T) {
	cfg, dsn := newADLTestConfig(t)
	app := newADLTestApp(t, cfg)

	const traceID = "8bf92f3577b34da6a3ce929d0e0e473a"

	resp := postEvaluation(t, app, map[string]string{
		"traceparent": "00-" + traceID + "-00f067aa0ba902b7-01",
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, 1, countDecisionsByTraceID(t, dsn, traceID))
}

// TestADLCompliance_SearchEventNamesAndErrorStatus covers the three Search API event names (§3.3.4) and
// the Error status path (§3.3.6): these fixtures carry no PIP enumeration data, so every search endpoint
// genuinely fails to evaluate rather than merely denying — the case status=Error is reserved for.
func TestADLCompliance_SearchEventNamesAndErrorStatus(t *testing.T) {
	testCases := []struct {
		name      string
		path      string
		body      string
		eventName string
		traceID   string
	}{
		{
			name:      "search subject",
			path:      "/authzen/v1/search/subject",
			body:      `{"subject":{"type":"doelbinding"},"action":{"name":"can_update"},"resource":{"type":"service","id":"x"}}`,
			eventName: "adl.search_subject",
			traceID:   "1bf92f3577b34da6a3ce929d0e0e47a1",
		},
		{
			name:      "search action",
			path:      "/authzen/v1/search/action",
			body:      `{"subject":{"type":"doelbinding","id":"subsidies"},"resource":{"type":"service","id":"x"}}`,
			eventName: "adl.search_action",
			traceID:   "2bf92f3577b34da6a3ce929d0e0e47a2",
		},
		{
			name:      "search resource",
			path:      "/authzen/v1/search/resource",
			body:      `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_update"},"resource":{"type":"service"}}`,
			eventName: "adl.search_resource",
			traceID:   "3bf92f3577b34da6a3ce929d0e0e47a3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, dsn := newADLTestConfig(t)
			app := newADLTestApp(t, cfg)

			req := httptest.NewRequest(http.MethodPost, tc.path, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("traceparent", "00-"+tc.traceID+"-00f067aa0ba902b7-01")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			row := queryDecisionByTraceID(t, dsn, tc.traceID)
			assert.Equal(t, tc.eventName, row.EventName)
			assert.Equal(t, "Error", row.Status, "§3.3.6: an evaluation failure MUST be Error, not a denial")
		})
	}
}

// TestADLCompliance_DenialIsNotFailure covers the other half of §3.3.6: a denied request is still a
// successful decision — Status must be Ok, never Error.
func TestADLCompliance_DenialIsNotFailure(t *testing.T) {
	cfg, dsn := newADLTestConfig(t)
	app := newADLTestApp(t, cfg)

	const traceID = "9bf92f3577b34da6a3ce929d0e0e473b"
	const deniedSubject = `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"GET"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`

	req := httptest.NewRequest(http.MethodPost, "/authzen/v1/evaluation", bytes.NewReader([]byte(deniedSubject)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("traceparent", "00-"+traceID+"-00f067aa0ba902b7-01")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	row := queryDecisionByTraceID(t, dsn, traceID)
	require.Contains(t, string(row.Body), `"not-authorized"`, "sanity check: this must genuinely be a denial")
	assert.Equal(t, "Ok", row.Status, "§3.3.6: a denial MUST still be Status Ok — Error is reserved for evaluation failures")
}

// TestADLCompliance_BodyEmbeddedTraceParentFallback covers Logius ADL §4.1.1.1 Example 10: with no
// traceparent HTTP header, a traceparent embedded in the AuthZEN request body's context object is used
// instead of an unrelated fresh trace, and its span_id becomes the record's parent — exactly like the
// header path.
func TestADLCompliance_BodyEmbeddedTraceParentFallback(t *testing.T) {
	cfg, dsn := newADLTestConfig(t)
	app := newADLTestApp(t, cfg)

	const bodyTraceID = "5bf92f3577b34da6a3ce929d0e0e4736"
	const bodySpanID = "00f067aa0ba902b7"
	const bodyTraceParent = "00-" + bodyTraceID + "-" + bodySpanID + "-01"

	body := `{` + goodEvaluationItem + `,"context":{"traceparent":"` + bodyTraceParent + `"}}`

	req := httptest.NewRequest(http.MethodPost, "/authzen/v1/evaluation", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	// Deliberately no traceparent header: the body-embedded fallback is what's under test.

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	row := queryDecisionByTraceID(t, dsn, bodyTraceID)
	require.NotNil(t, row.ParentSpanID, "the body-embedded span_id MUST become the parent, same as the header path")
	assert.Equal(t, bodySpanID, *row.ParentSpanID)
}
