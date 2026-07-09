package adl

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// memSink is an in-memory Sink capturing emitted records.
type memSink struct {
	mu   sync.Mutex
	recs []*Record
}

func (m *memSink) Emit(_ context.Context, rec *Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recs = append(m.recs, rec)
	return nil
}

func (m *memSink) records() []*Record {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*Record(nil), m.recs...)
}

func newAuthRecord() *authlog.AuthRecord {
	t := time.UnixMilli(1757240058042).UTC()
	dc := models.NewAttributeSet()
	dc.AddAttribute("message", "No signing authority")
	dc.AddAttribute("policy", "cedar/hr")
	dc.AddAttribute("policyHash", "6266d07750c44b4c9b05d0801b752c0ef884e4f6")

	return &authlog.AuthRecord{
		RequestTime:     &t,
		Principal:       models.NewEntity("user", "alice", nil),
		Action:          models.NewEntity(models.EntityTypeName, "approve", nil),
		Resource:        models.NewEntity("holiday-request", "446epbc8y7", models.NewAttributeSet(models.NewAttribute("employee", "bob"))),
		Decision:        false,
		DecisionContext: dc,
		TraceParent:     "00-28dbeec32e77635cc19bc3204ec56c41-893e1b2ac52d712f-01",
		RequestContext:  models.NewAttributeSet(models.NewAttribute("traceparent", "00-28dbeec32e77635cc19bc3204ec56c41-dec5220770f8f4f4-01")),
	}
}

func TestLog_RecordShape(t *testing.T) {
	t.Parallel()

	sink := &memSink{}
	l, err := New(Config{Level: 1, Resource: map[string]any{"service.name": "pdp-test"}}, sink)
	require.NoError(t, err)

	require.NoError(t, l.Log(context.Background(), true, newAuthRecord()))

	recs := sink.records()
	require.Len(t, recs, 1, "exactly one record per evaluation")
	rec := recs[0]

	assert.Regexp(t, hex32, rec.TraceID)
	assert.Regexp(t, hex16, rec.SpanID)
	assert.Equal(t, "28dbeec32e77635cc19bc3204ec56c41", rec.TraceID)
	assert.Equal(t, "893e1b2ac52d712f", rec.ParentSpanID)
	assert.Equal(t, EventAccessEvaluation, rec.EventName)
	assert.Equal(t, uint64(1757240058042), rec.Timestamp)

	// A deny is a valid outcome: status Ok, never Error.
	assert.Equal(t, StatusOk, rec.Status)

	assert.Equal(t, map[string]any{"service.name": "pdp-test"}, rec.Resource)

	// Level 1: full AuthZEN request and response in body.
	req, ok := rec.Body[KeyRequest].(map[string]any)
	require.True(t, ok, "body must carry adl.core.request")
	assert.Equal(t, map[string]any{"type": "user", "id": "alice"}, req["subject"])
	assert.Equal(t, map[string]any{"name": "approve"}, req["action"])
	assert.Equal(t, map[string]any{
		"type": "holiday-request", "id": "446epbc8y7",
		"properties": map[string]any{"employee": "bob"},
	}, req["resource"])
	assert.Equal(t, map[string]any{"traceparent": "00-28dbeec32e77635cc19bc3204ec56c41-dec5220770f8f4f4-01"}, req["context"])

	resp, ok2 := rec.Body[KeyResponse].(map[string]any)
	require.True(t, ok2, "body must carry adl.core.response")
	assert.Equal(t, false, resp["decision"])

	// Level 1: no adl.core.* source references in attributes (location rule: body XOR attributes).
	assert.NotContains(t, rec.Attributes, KeyPolicies)
	assert.NotContains(t, rec.Attributes, KeyRequest)
	assert.NotContains(t, rec.Attributes, KeyResponse)
	assert.NotContains(t, rec.Body, KeyPolicies)
}

func TestLog_ErroredEvaluation(t *testing.T) {
	t.Parallel()

	sink := &memSink{}
	l, err := New(Config{Level: 1}, sink)
	require.NoError(t, err)

	ar := newAuthRecord()
	ar.Errored = true

	require.NoError(t, l.Log(context.Background(), true, ar))

	recs := sink.records()
	require.Len(t, recs, 1, "an errored evaluation still yields exactly one record")
	assert.Equal(t, StatusError, recs[0].Status)
	assert.NotContains(t, recs[0].Body, KeyResponse, "response may be omitted when status is Error")
}

// A permit with "ja, mits" duties must carry those obligations in the logged
// response context, not just the bare decision: the ADL record has to reproduce
// what the PDP returned to the PEP (auth_process.go stores them in the decision
// context under "obligations").
func TestLog_ResponseCarriesObligations(t *testing.T) {
	t.Parallel()

	sink := &memSink{}
	l, err := New(Config{Level: 1}, sink)
	require.NoError(t, err)

	obs := []map[string]any{
		{"id": "https://standaarden.overheid.nl/odrl-geo-nl/filterFeatures",
			"properties": map[string]any{"property": "gevoeligheidswaarde", "operator": "lt", "value": 4}},
	}
	ar := newAuthRecord()
	ar.Decision = true
	ar.DecisionContext.AddAttribute("obligations", obs)

	require.NoError(t, l.Log(context.Background(), true, ar))

	recs := sink.records()
	require.Len(t, recs, 1)
	resp, ok := recs[0].Body[KeyResponse].(map[string]any)
	require.True(t, ok, "body must carry adl.core.response")
	ctx, ok := resp["context"].(map[string]any)
	require.True(t, ok, "response must carry a context")
	assert.Equal(t, obs, ctx["obligations"], "obligations must be logged in the response context")
}

func TestLog_EventNames(t *testing.T) {
	t.Parallel()

	sink := &memSink{}
	l, err := New(Config{Level: 1}, sink)
	require.NoError(t, err)

	for _, name := range []string{
		EventAccessEvaluation, EventAccessEvaluations,
		EventSearchSubject, EventSearchAction, EventSearchResource,
	} {
		ar := newAuthRecord()
		ar.TraceParent = "" // force fresh ids so each record is unique.
		ar.EventName = name
		require.NoError(t, l.Log(context.Background(), true, ar))
	}

	recs := sink.records()
	require.Len(t, recs, 5)
	for i, name := range []string{
		EventAccessEvaluation, EventAccessEvaluations,
		EventSearchSubject, EventSearchAction, EventSearchResource,
	} {
		assert.Equal(t, name, recs[i].EventName)
	}
}

func TestLog_Levels(t *testing.T) {
	t.Parallel()

	info := informationProviderFunc(func(_ context.Context, _ *authlog.AuthRecord) map[string]any {
		return map[string]any{"can-sign-api": map[string]any{"span_id": "836ff5286112f460"}}
	})

	testCases := []struct {
		level    int
		wantPol  bool
		wantInfo bool
		wantConf bool
	}{
		{level: 1},
		{level: 2, wantPol: true},
		{level: 3, wantPol: true, wantInfo: true},
		{level: 4, wantPol: true, wantInfo: true, wantConf: true},
	}

	for _, tc := range testCases {
		sink := &memSink{}
		l, err := New(Config{
			Level:         tc.level,
			Information:   info,
			Engine:        "cedar",
			EngineVersion: "1.2.3",
			ConfigHash:    "cafebabe",
		}, sink)
		require.NoError(t, err)

		require.NoError(t, l.Log(context.Background(), true, newAuthRecord()))
		recs := sink.records()
		require.Len(t, recs, 1, "level %d", tc.level)
		rec := recs[0]

		if tc.wantPol {
			// Level >= 2: policy version reference in attributes (policy key + hash).
			assert.Equal(t, map[string]any{"cedar/hr": "6266d07750c44b4c9b05d0801b752c0ef884e4f6"},
				rec.Attributes[KeyPolicies], "level %d", tc.level)
			assert.NotContains(t, rec.Body, KeyPolicies, "level %d: policies reference must not also be in body", tc.level)
		} else {
			assert.NotContains(t, rec.Attributes, KeyPolicies, "level %d", tc.level)
		}

		if tc.wantInfo {
			assert.Equal(t, map[string]any{"can-sign-api": map[string]any{"span_id": "836ff5286112f460"}},
				rec.Attributes[KeyInformation], "level %d", tc.level)
		} else {
			assert.NotContains(t, rec.Attributes, KeyInformation, "level %d", tc.level)
		}

		if tc.wantConf {
			assert.Equal(t, map[string]any{"engine": "cedar", "version": "1.2.3", "config_hash": "cafebabe"},
				rec.Body[KeyConfiguration], "level %d", tc.level)
		} else {
			assert.NotContains(t, rec.Body, KeyConfiguration, "level %d", tc.level)
		}
	}
}

func TestLog_WALAppend(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "adl", "wal.jsonl")
	l, err := New(Config{Level: 2, WALPath: path})
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	require.NoError(t, l.Log(context.Background(), true, newAuthRecord()))
	ar := newAuthRecord()
	ar.TraceParent = ""
	require.NoError(t, l.Log(context.Background(), true, ar))

	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	var lines []*Record
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		rec := &Record{}
		require.NoError(t, json.Unmarshal(scanner.Bytes(), rec))
		lines = append(lines, rec)
	}
	require.NoError(t, scanner.Err())

	require.Len(t, lines, 2, "one JSONL line per evaluation")
	assert.Equal(t, "28dbeec32e77635cc19bc3204ec56c41", lines[0].TraceID)
	assert.Regexp(t, hex32, lines[1].TraceID)
	assert.Empty(t, lines[1].ParentSpanID)
	assert.NotEqual(t, lines[0].SpanID, lines[1].SpanID)
}

func TestEmit_Idempotent(t *testing.T) {
	t.Parallel()

	sink := &memSink{}
	l, err := New(Config{Level: 1}, sink)
	require.NoError(t, err)

	rec := &Record{TraceID: "28dbeec32e77635cc19bc3204ec56c41", SpanID: "5e3c8a4f9b2d1e07", EventName: EventAccessEvaluation, Timestamp: 1, Status: StatusOk}

	l.Emit(context.Background(), rec)
	l.Emit(context.Background(), rec) // duplicate delivery.
	l.Emit(context.Background(), &Record{TraceID: rec.TraceID, SpanID: "17c59821784ee492", EventName: EventAccessEvaluation, Timestamp: 2, Status: StatusOk})

	recs := sink.records()
	require.Len(t, recs, 2, "duplicate (trace_id, span_id) must not be emitted twice")
}

func TestReplay_IsIdempotent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "wal.jsonl")
	sink := &memSink{}
	l, err := New(Config{Level: 1, WALPath: path}, sink)
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	require.NoError(t, l.Log(context.Background(), true, newAuthRecord()))
	require.Len(t, sink.records(), 1)

	// Replaying the WAL must not duplicate the already-flushed record.
	require.NoError(t, l.Replay(context.Background()))
	assert.Len(t, sink.records(), 1)

	// A fresh logger instance (e.g. after a crash) re-emits the record exactly once.
	sink2 := &memSink{}
	l2, err := New(Config{Level: 1, WALPath: path}, sink2)
	require.NoError(t, err)
	defer func() { _ = l2.Close() }()

	require.NoError(t, l2.Replay(context.Background()))
	require.NoError(t, l2.Replay(context.Background()))
	assert.Len(t, sink2.records(), 1)
}

// informationProviderFunc adapts a function to the InformationProvider interface.
type informationProviderFunc func(ctx context.Context, ar *authlog.AuthRecord) map[string]any

func (f informationProviderFunc) Information(ctx context.Context, ar *authlog.AuthRecord) map[string]any {
	return f(ctx, ar)
}
