package adl

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci/opensearch"
)

// fakeOSLogger captures the LogRecords an openSearchSink writes.
type fakeOSLogger struct{ recs []opensearch.LogRecord }

func (f *fakeOSLogger) Log(_ context.Context, _ bool, rec opensearch.LogRecord) error {
	f.recs = append(f.recs, rec)
	return nil
}
func (f *fakeOSLogger) LogBulk(_ context.Context, _ bool, recs ...opensearch.LogRecord) error {
	f.recs = append(f.recs, recs...)
	return nil
}
func (f *fakeOSLogger) CreateIndex(context.Context, string, int, int) error { return nil }
func (f *fakeOSLogger) DeleteIndexes(context.Context, ...string) error      { return nil }

// I1: the OpenSearch sink writes with a deterministic document ID (trace_id:span_id) so
// redelivery overwrites rather than duplicates.
func TestOpenSearchSink_DeterministicDocumentID(t *testing.T) {
	fake := &fakeOSLogger{}
	orig := newOpenSearchLogger
	newOpenSearchLogger = func(_, _ string, _ []string) (opensearch.Logger, error) { return fake, nil }
	defer func() { newOpenSearchLogger = orig }()

	sink, err := NewOpenSearchSink("adl", "u", "p", "https://opensearch:9200")
	require.NoError(t, err)

	rec := richRecord()
	require.NoError(t, sink.Emit(context.Background(), rec))
	require.NoError(t, sink.Emit(context.Background(), rec)) // redelivery.

	require.Len(t, fake.recs, 2)
	for _, r := range fake.recs {
		assert.Equal(t, rec.IdempotencyKey(), r.ID, "document ID must be the idempotency key")
		assert.Equal(t, "0af7651916cd43dd8448eb211c80319c:b7ad6b7169203331", r.ID)
	}
}

// F18: adl.fsc.transaction_id is set exactly when the request crossed an FSC in/outway.
func TestBuild_FSCTransactionID(t *testing.T) {
	t.Parallel()

	sink := &memSink{}
	l, err := New(Config{Level: 1}, sink)
	require.NoError(t, err)

	// FSC path: transaction id present -> attribute set.
	ar := newAuthRecord()
	ar.FSCTransactionID = "0190f2c0-7a11-7000-8000-abcabcabcabc"
	require.NoError(t, l.Log(context.Background(), true, ar))

	// AuthZEN path: no transaction id -> attribute absent.
	az := newAuthRecord()
	az.TraceParent = "" // fresh ids so it is not deduplicated against the first record.
	require.NoError(t, l.Log(context.Background(), true, az))

	recs := sink.records()
	require.Len(t, recs, 2)
	assert.Equal(t, "0190f2c0-7a11-7000-8000-abcabcabcabc", recs[0].Attributes[KeyTransactionID],
		"FSC request must carry adl.fsc.transaction_id")
	assert.NotContains(t, recs[1].Attributes, KeyTransactionID,
		"AuthZEN request must NOT carry adl.fsc.transaction_id")
}

// F15/F20: a policy without a resolvable hash must NOT produce a {key:key} pseudo-reference;
// the attribute is omitted and the record is degraded with a warning.
func TestBuild_PoliciesRequireResolvableHash(t *testing.T) {
	t.Parallel()

	var logbuf strings.Builder
	logger := slog.New(slog.NewTextHandler(&logbuf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	sink := &memSink{}
	l, err := New(Config{Level: 2, Logger: logger}, sink)
	require.NoError(t, err)

	ar := newAuthRecord()
	// A policy key is known, but no hash: not a resolvable version reference.
	ar.DecisionContext = models.NewAttributeSet()
	ar.DecisionContext.AddAttribute("policy", "cedar/hr")

	require.NoError(t, l.Log(context.Background(), true, ar))

	recs := sink.records()
	require.Len(t, recs, 1)
	assert.NotContains(t, recs[0].Attributes, KeyPolicies,
		"no {key:key} fallback: adl.core.policies must be omitted without a resolvable hash")
	assert.Contains(t, logbuf.String(), "omitting adl.core.policies",
		"the degraded record must be flagged with a warning")

	// With a hash, the reference IS emitted (parallel engine work fills the hash).
	sink2 := &memSink{}
	l2, err := New(Config{Level: 2}, sink2)
	require.NoError(t, err)
	require.NoError(t, l2.Log(context.Background(), true, newAuthRecord()))
	recs2 := sink2.records()
	require.Len(t, recs2, 1)
	assert.Equal(t, map[string]any{"cedar/hr": "6266d07750c44b4c9b05d0801b752c0ef884e4f6"},
		recs2[0].Attributes[KeyPolicies])
}

// IM2: a data-subject reference supplied on the request context is promoted to attributes.
func TestBuild_DataSubjectPassthrough(t *testing.T) {
	t.Parallel()

	sink := &memSink{}
	l, err := New(Config{Level: 1}, sink)
	require.NoError(t, err)

	ar := newAuthRecord()
	ar.RequestContext = models.NewAttributeSet(
		models.NewAttribute(KeyDataSubjectID, "bsn:999999990"),
		models.NewAttribute(KeyDataSubjectType, "natuurlijk_persoon"),
	)
	require.NoError(t, l.Log(context.Background(), true, ar))

	recs := sink.records()
	require.Len(t, recs, 1)
	assert.Equal(t, "bsn:999999990", recs[0].Attributes[KeyDataSubjectID])
	assert.Equal(t, "natuurlijk_persoon", recs[0].Attributes[KeyDataSubjectType])

	// Absent on the request context -> absent on the record.
	sink2 := &memSink{}
	l2, err := New(Config{Level: 1}, sink2)
	require.NoError(t, err)
	require.NoError(t, l2.Log(context.Background(), true, newAuthRecord()))
	recs2 := sink2.records()
	require.Len(t, recs2, 1)
	assert.NotContains(t, recs2[0].Attributes, KeyDataSubjectID)
}

// I1 (OTLP logs path): the record's idempotency key is carried as log.record.uid so a
// backend can deduplicate redelivered records.
func TestEncodeOTLPLog_CarriesIdempotencyKey(t *testing.T) {
	t.Parallel()

	rec := richRecord()
	payload, err := EncodeOTLPLog(rec, "adl")
	require.NoError(t, err)

	_, labels := decodeOTLPLine(t, payload)
	assert.Equal(t, rec.IdempotencyKey(), labels[LabelLogRecordUID])
	assert.Equal(t, "0af7651916cd43dd8448eb211c80319c:b7ad6b7169203331", labels[LabelLogRecordUID])
}

// B1: cleartext endpoints are rejected unless loopback or an explicit insecure opt-out.
func TestRequireSecureEndpoint(t *testing.T) {
	defer func() { AllowInsecureTransport = false }()

	AllowInsecureTransport = false
	assert.Error(t, requireSecureEndpoint("http://collector:4318"), "cleartext non-loopback must be rejected")
	assert.NoError(t, requireSecureEndpoint("https://collector:4318"), "TLS endpoint allowed")
	assert.NoError(t, requireSecureEndpoint("http://localhost:4318"), "loopback allowed")
	assert.NoError(t, requireSecureEndpoint("http://127.0.0.1:4318"), "loopback IP allowed")
	assert.NoError(t, requireSecureEndpoint("http://[::1]:4318"), "loopback IPv6 allowed")
	assert.NoError(t, requireSecureEndpoint("stdout"), "non-URL target allowed")
	assert.NoError(t, requireSecureEndpoint(""), "empty allowed")

	AllowInsecureTransport = true
	assert.NoError(t, requireSecureEndpoint("http://collector:4318"), "explicit opt-out permits cleartext")
}

// B1: the sink constructors enforce requireSecureEndpoint.
func TestSinkConstructors_RejectCleartext(t *testing.T) {
	defer func() { AllowInsecureTransport = false }()
	AllowInsecureTransport = false

	_, err := NewOTLPLogSink(OTLPLogConfig{Endpoint: "http://collector:4318"})
	assert.Error(t, err, "OTLP log sink must reject cleartext")

	_, err = NewOpenSearchSink("idx", "u", "p", "http://opensearch:9200")
	assert.Error(t, err, "OpenSearch sink must reject cleartext")
}

// richRecord's marshalling is stable enough to assert the encoded uid stays a plain string.
func TestEncodeOTLPLog_UIDIsString(t *testing.T) {
	t.Parallel()
	rec := richRecord()
	payload, err := EncodeOTLPLog(rec, "adl")
	require.NoError(t, err)
	var probe map[string]any
	require.NoError(t, json.Unmarshal(payload, &probe))
}
