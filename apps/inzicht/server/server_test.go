package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/inzicht/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/cloudevents"
)

// Bearer tokens wired into the test service (simulation stand-in for OAuth2/mTLS).
const (
	tokRVIG      = "tok-rvig"      // verstrekker "rvig".
	tokPolitie   = "tok-politie"   // verstrekker "politie".
	tokBeheerder = "tok-beheerder" // beheerder "afnemer-admin".
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func evalRecord(traceID, afnemer, doel string, permit bool, ts time.Time) adl.Record {
	return adl.Record{
		TraceID:   traceID,
		SpanID:    "0000000000000001",
		EventName: adl.EventAccessEvaluation,
		Timestamp: uint64(ts.UnixMilli()),
		Status:    adl.StatusOk,
		Attributes: map[string]any{
			adl.KeyTransactionID: "tx-" + traceID,
			adl.KeyPolicies:      map[string]any{"brp-2026": "v1"},
		},
		Body: map[string]any{
			adl.KeyRequest: map[string]any{
				"subject": map[string]any{"id": afnemer},
				"action":  map[string]any{"name": "read"},
				"context": map[string]any{"purpose": doel},
			},
			adl.KeyResponse: map[string]any{"decision": permit},
		},
	}
}

// newTestConfig writes a synthetic WAL (2 afnemers x 3 doelen) and returns a config
// pointing at it, with the default bearer tokens configured.
func newTestConfig(t *testing.T) *config.Config {
	t.Helper()
	dir := t.TempDir()
	walPath := filepath.Join(dir, "adl.jsonl")

	f, err := os.Create(walPath)
	if err != nil {
		t.Fatal(err)
	}
	enc := json.NewEncoder(f)
	day := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	n := 0
	for _, a := range []string{"gemeente-amsterdam", "gemeente-utrecht"} {
		for _, d := range []string{"opsporing", "huisvestingswet", "woz"} {
			for i := 0; i < 6; i++ {
				n++
				_ = enc.Encode(evalRecord(tid(n), a, d, true, day))
			}
			for i := 0; i < 2; i++ {
				n++
				_ = enc.Encode(evalRecord(tid(n), a, d, false, day))
			}
		}
	}
	_ = f.Close()

	cfg := &config.Config{}
	cfg.ADL.Path = walPath
	cfg.ADL.Level = 1
	cfg.Inzicht.K = 5
	cfg.Inzicht.Bucket = "day"
	cfg.Inzicht.Source = "nl.overheid.ftv.inzicht.test"
	cfg.Inzicht.AuthTokens = strings.Join([]string{
		tokRVIG + ":verstrekker:rvig",
		tokPolitie + ":verstrekker:politie",
		tokBeheerder + ":beheerder:afnemer-admin",
	}, ",")
	return cfg
}

// newTestService returns a service backed by an in-memory KV store and the synthetic
// WAL, with authentication enabled and the default tokens configured.
func newTestService(t *testing.T) (*Service, string) {
	t.Helper()
	cfg := newTestConfig(t)
	svc := serviceFromConfig(t, cfg)
	return svc, cfg.ADL.Path
}

func serviceFromConfig(t *testing.T, cfg *config.Config) *Service {
	t.Helper()
	svc, err := NewService(context.Background(), cfg, testLogger())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	return svc
}

func tid(n int) string {
	const hex = "0123456789abcdef"
	b := []byte(strings.Repeat("0", 32))
	b[31] = hex[n%16]
	b[30] = hex[(n/16)%16]
	return string(b)
}

// do issues a request carrying the given bearer token (empty token = anonymous).
func do(t *testing.T, svc *Service, token, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != nil {
		buf, _ := json.Marshal(body)
		r = httptest.NewRequest(method, target, bytes.NewReader(buf))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	svc.Handler().ServeHTTP(w, r)
	return w
}

func TestStatistiekenEndpoint(t *testing.T) {
	svc, _ := newTestService(t)
	w := do(t, svc, tokRVIG, http.MethodGet, "/v1/statistieken?bucket=day", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
	var resp StatisticsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Statistieken) != 6 {
		t.Fatalf("got %d buckets, want 6", len(resp.Statistieken))
	}
	for _, c := range resp.Statistieken {
		if c.Permit != 6 || c.Deny != 2 {
			t.Fatalf("bucket %+v unexpected", c)
		}
	}
}

func TestApprovalFlow(t *testing.T) {
	svc, _ := newTestService(t)

	// The verstrekker (rvig) creates a pending request for one doel.
	w := do(t, svc, tokRVIG, http.MethodPost, "/v1/inzicht/verzoeken", createRequest{
		Doel: "opsporing",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body=%s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("no id returned")
	}
	if created["status"] != "pending" {
		t.Fatalf("status = %v, want pending", created["status"])
	}
	// Verstrekker identity comes from the token, not the request body.
	if created["verstrekker"] != "rvig" {
		t.Fatalf("verstrekker = %v, want rvig (from token)", created["verstrekker"])
	}

	// Resultaat before approval -> 403.
	w = do(t, svc, tokRVIG, http.MethodGet, "/v1/inzicht/verzoeken/"+id+"/resultaat", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("resultaat before approve = %d, want 403", w.Code)
	}

	// List (beheerder only) shows the pending request.
	w = do(t, svc, tokBeheerder, http.MethodGet, "/v1/inzicht/verzoeken", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), id) {
		t.Fatalf("list did not contain %s: %d %s", id, w.Code, w.Body.String())
	}

	// Approve (beheerder).
	w = do(t, svc, tokBeheerder, http.MethodPost, "/v1/inzicht/verzoeken/"+id+"/approve",
		decideRequest{DecidedBy: "beheerder"})
	if w.Code != http.StatusOK {
		t.Fatalf("approve = %d, body=%s", w.Code, w.Body.String())
	}

	// Resultaat after approval -> the 16 records for doel=opsporing (both afnemers).
	w = do(t, svc, tokRVIG, http.MethodGet, "/v1/inzicht/verzoeken/"+id+"/resultaat", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("resultaat = %d, body=%s", w.Code, w.Body.String())
	}
	var result struct {
		Verwerkingen []Verwerking `json:"verwerkingen"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	if len(result.Verwerkingen) != 16 { // 2 afnemers x 8 records for opsporing.
		t.Fatalf("got %d verwerkingen, want 16", len(result.Verwerkingen))
	}
	// Location rule: references present, no raw request/response payloads.
	vw := result.Verwerkingen[0]
	if vw.TraceID == "" || vw.Decision == nil || vw.Doel != "opsporing" {
		t.Fatalf("projection missing fields: %+v", vw)
	}
	if vw.References[adl.KeyTransactionID] == nil {
		t.Fatalf("expected transaction_id reference, got %+v", vw.References)
	}
}

func TestDenyForbidsResult(t *testing.T) {
	svc, _ := newTestService(t)
	w := do(t, svc, tokRVIG, http.MethodPost, "/v1/inzicht/verzoeken", createRequest{Doel: "woz"})
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	w = do(t, svc, tokBeheerder, http.MethodPost, "/v1/inzicht/verzoeken/"+id+"/deny",
		decideRequest{Reason: "onvoldoende grondslag"})
	if w.Code != http.StatusOK {
		t.Fatalf("deny = %d", w.Code)
	}

	w = do(t, svc, tokRVIG, http.MethodGet, "/v1/inzicht/verzoeken/"+id+"/resultaat", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("resultaat after deny = %d, want 403", w.Code)
	}
}

func TestInzichtAccessLogged(t *testing.T) {
	svc, walPath := newTestService(t)

	before := countEvents(t, walPath, EventInzichtAccess)

	// A statistics call (verstrekker) and a list call (beheerder) each log one record.
	_ = do(t, svc, tokRVIG, http.MethodGet, "/v1/statistieken", nil)
	_ = do(t, svc, tokBeheerder, http.MethodGet, "/v1/inzicht/verzoeken", nil)

	after := countEvents(t, walPath, EventInzichtAccess)
	if after-before != 2 {
		t.Fatalf("access records logged = %d, want 2", after-before)
	}
	if EventInzichtAccess != adl.EventSearchResource {
		t.Fatalf("unexpected access event_name %q", EventInzichtAccess)
	}
}

// TestAuthClosedByDefault: with no tokens configured and auth enabled, every
// protected route is refused with 401 (fail-closed), while /healthz stays open.
func TestAuthClosedByDefault(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Inzicht.AuthTokens = ""
	svc := serviceFromConfig(t, cfg)

	for _, target := range []string{"/v1/statistieken", "/v1/inzicht/verzoeken"} {
		if w := do(t, svc, "", http.MethodGet, target, nil); w.Code != http.StatusUnauthorized {
			t.Fatalf("GET %s = %d, want 401", target, w.Code)
		}
	}
	if w := do(t, svc, "", http.MethodGet, "/healthz", nil); w.Code != http.StatusOK {
		t.Fatalf("healthz = %d, want 200", w.Code)
	}
}

// TestUnknownAndMissingTokenRejected: bad or absent bearer token -> 401.
func TestUnknownAndMissingTokenRejected(t *testing.T) {
	svc, _ := newTestService(t)
	if w := do(t, svc, "", http.MethodGet, "/v1/statistieken", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("no token = %d, want 401", w.Code)
	}
	if w := do(t, svc, "bogus", http.MethodGet, "/v1/statistieken", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("bad token = %d, want 401", w.Code)
	}
}

// TestAuthDisabled: the local-play opt-out accepts an X-Verstrekker identity.
func TestAuthDisabled(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Inzicht.AuthTokens = ""
	cfg.Inzicht.AuthDisabled = true
	svc := serviceFromConfig(t, cfg)

	r := httptest.NewRequest(http.MethodPost, "/v1/inzicht/verzoeken", bytes.NewReader(mustJSON(createRequest{Doel: "woz"})))
	r.Header.Set("X-Verstrekker", "local-dev")
	w := httptest.NewRecorder()
	svc.Handler().ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("create (auth disabled) = %d, body=%s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if created["verstrekker"] != "local-dev" {
		t.Fatalf("verstrekker = %v, want local-dev", created["verstrekker"])
	}
}

// TestApproveRequiresBeheerder: a verstrekker cannot self-approve.
func TestApproveRequiresBeheerder(t *testing.T) {
	svc, _ := newTestService(t)
	w := do(t, svc, tokRVIG, http.MethodPost, "/v1/inzicht/verzoeken", createRequest{Doel: "opsporing"})
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	if w := do(t, svc, tokRVIG, http.MethodPost, "/v1/inzicht/verzoeken/"+id+"/approve", nil); w.Code != http.StatusForbidden {
		t.Fatalf("verstrekker self-approve = %d, want 403", w.Code)
	}
	// The beheerder cannot create a verzoek (no verstrekker identity).
	if w := do(t, svc, tokBeheerder, http.MethodPost, "/v1/inzicht/verzoeken", createRequest{Doel: "woz"}); w.Code != http.StatusForbidden {
		t.Fatalf("beheerder create = %d, want 403", w.Code)
	}
}

// TestResultaatOnlyForSubmitter: another verstrekker cannot read the result.
func TestResultaatOnlyForSubmitter(t *testing.T) {
	svc, _ := newTestService(t)
	w := do(t, svc, tokRVIG, http.MethodPost, "/v1/inzicht/verzoeken", createRequest{Doel: "opsporing"})
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	_ = do(t, svc, tokBeheerder, http.MethodPost, "/v1/inzicht/verzoeken/"+id+"/approve", nil)

	// politie (a different verstrekker) is refused.
	if w := do(t, svc, tokPolitie, http.MethodGet, "/v1/inzicht/verzoeken/"+id+"/resultaat", nil); w.Code != http.StatusForbidden {
		t.Fatalf("other verstrekker resultaat = %d, want 403", w.Code)
	}
	// the submitting verstrekker succeeds.
	if w := do(t, svc, tokRVIG, http.MethodGet, "/v1/inzicht/verzoeken/"+id+"/resultaat", nil); w.Code != http.StatusOK {
		t.Fatalf("submitter resultaat = %d, want 200", w.Code)
	}
}

// TestResultaatAfnemerScope: the result is limited to the approved afnemer scope.
func TestResultaatAfnemerScope(t *testing.T) {
	svc, _ := newTestService(t)
	w := do(t, svc, tokRVIG, http.MethodPost, "/v1/inzicht/verzoeken", createRequest{
		Doel:    "opsporing",
		Afnemer: "gemeente-amsterdam",
	})
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	_ = do(t, svc, tokBeheerder, http.MethodPost, "/v1/inzicht/verzoeken/"+id+"/approve", nil)

	w = do(t, svc, tokRVIG, http.MethodGet, "/v1/inzicht/verzoeken/"+id+"/resultaat", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("resultaat = %d, body=%s", w.Code, w.Body.String())
	}
	var result struct {
		Verwerkingen []Verwerking `json:"verwerkingen"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	if len(result.Verwerkingen) != 8 { // one afnemer x 8 records for opsporing.
		t.Fatalf("got %d verwerkingen, want 8 (scoped to one afnemer)", len(result.Verwerkingen))
	}
	for _, vw := range result.Verwerkingen {
		if vw.Afnemer != "gemeente-amsterdam" {
			t.Fatalf("result leaked afnemer %q outside approved scope", vw.Afnemer)
		}
	}
}

func TestStatisticsPush(t *testing.T) {
	svc, _ := newTestService(t)

	// Aggregate over a window that covers the synthetic day.
	now = func() time.Time { return time.Date(2026, 7, 1, 23, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = time.Now })
	svc.cfg.Inzicht.PushWindow = 48 * time.Hour

	received := make(chan []byte, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		if !cloudevents.IsStructured(r.Header.Get("Content-Type")) {
			t.Errorf("content-type not structured: %s", r.Header.Get("Content-Type"))
		}
		received <- buf.Bytes()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	svc.pushOnce(context.Background(), []string{ts.URL})

	select {
	case body := <-received:
		ev, err := cloudevents.ParseStructured(body)
		if err != nil {
			t.Fatalf("parse cloudevent: %v", err)
		}
		if ev.Type != StatisticsEventType {
			t.Fatalf("event type = %q, want %q", ev.Type, StatisticsEventType)
		}
		var data StatisticsResponse
		if err := json.Unmarshal(ev.Data, &data); err != nil {
			t.Fatalf("parse data: %v", err)
		}
		if len(data.Statistieken) != 6 {
			t.Fatalf("pushed %d buckets, want 6", len(data.Statistieken))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no event received")
	}
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func countEvents(t *testing.T, walPath, eventName string) int {
	t.Helper()
	f, err := os.Open(walPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	count := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var rec adl.Record
		if json.Unmarshal(sc.Bytes(), &rec) == nil && rec.EventName == eventName {
			count++
		}
	}
	return count
}
