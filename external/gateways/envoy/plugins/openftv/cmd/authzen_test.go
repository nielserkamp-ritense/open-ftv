package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	types "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
)

// Verifies trace context is propagated correctly across the two hops Envoy sits on: the call to the PDP,
// and the call it authorizes onward to the upstream service.

func newTestAuthServer(t *testing.T, pdpURL string) *authServer {
	t.Helper()

	return &authServer{
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		pdp:        pdpURL,
		timeout:    time.Second,
		httpClient: &http.Client{Timeout: time.Second},
		pep:        pep.New(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil))),
	}
}

func newCheckRequest(method string, headers map[string]string) *auth.CheckRequest {
	return &auth.CheckRequest{
		Attributes: &auth.AttributeContext{
			Request: &auth.AttributeContext_Request{
				Http: &auth.AttributeContext_HttpRequest{
					Method:  method,
					Scheme:  "https",
					Host:    "svc.internal",
					Path:    "/resource",
					Headers: headers,
				},
			},
		},
	}
}

func headerValue(headers []*corev3.HeaderValueOption, key string) (string, bool) {
	for _, h := range headers {
		if h.GetHeader().GetKey() == key {
			return h.GetHeader().GetValue(), true
		}
	}

	return "", false
}

func decisionPDP(t *testing.T, decision bool, checkReq func(r *http.Request)) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checkReq != nil {
			checkReq(r)
		}

		w.Header().Set("Content-Type", "application/json")

		if decision {
			_, _ = w.Write([]byte(`{"decision":true}`))
		} else {
			_, _ = w.Write([]byte(`{"decision":false}`))
		}
	}))
}

// TestCheck_Allowed_PropagatesTraceContextAcrossBothHops verifies §3.1 (both trace-context headers sent
// on every hop) and §3.2.2 (each hop gets its own fresh child span under the same trace_id, and
// tracestate is carried through unchanged).
func TestCheck_Allowed_PropagatesTraceContextAcrossBothHops(t *testing.T) {
	t.Parallel()

	const (
		incomingTraceID     = "4bf92f3577b34da6a3ce929d0e0e4736"
		incomingSpanID      = "00f067aa0ba902b7"
		incomingTraceParent = "00-" + incomingTraceID + "-" + incomingSpanID + "-01"
		incomingTraceState  = "vendorA=1,vendorB=2"
		fscTxnID            = "fsc-txn-abc123"
	)

	var pdpTraceParent, pdpTraceState, pdpFSCTxnID string

	pdp := decisionPDP(t, true, func(r *http.Request) {
		pdpTraceParent = r.Header.Get(models.HeaderTraceParent)
		pdpTraceState = r.Header.Get(models.HeaderTraceState)
		pdpFSCTxnID = r.Header.Get(models.HeaderFSCTransactionID)
	})
	defer pdp.Close()

	server := newTestAuthServer(t, pdp.URL)

	req := newCheckRequest(http.MethodGet, map[string]string{
		models.HeaderTraceParent:      incomingTraceParent,
		models.HeaderTraceState:       incomingTraceState,
		models.HeaderFSCTransactionID: fscTxnID,
	})

	resp, err := server.Check(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	ok, isOk := resp.HttpResponse.(*auth.CheckResponse_OkResponse)
	require.True(t, isOk, "expected an OkResponse for an allowed decision")

	// §3.1: both trace-context headers, on the hop to the PDP...
	pdpTC, ok2 := models.ParseTraceParent(pdpTraceParent)
	require.True(t, ok2, "PDP call must carry a valid traceparent")
	assert.Equal(t, incomingTraceID, pdpTC.TraceID, "§3.2.2: trace_id MUST be preserved")
	assert.NotEqual(t, incomingSpanID, pdpTC.SpanID, "§3.2.2: the PDP hop MUST mint its own child span")
	assert.Equal(t, incomingTraceState, pdpTraceState, "§3.2.2/W3C §3.5: tracestate MUST be carried through unchanged")
	assert.Equal(t, fscTxnID, pdpFSCTxnID)

	// ...and on the forwarded hop to the upstream service.
	upstreamTP, hasTP := headerValue(ok.OkResponse.Headers, models.HeaderTraceParent)
	require.True(t, hasTP, "§3.1: upstream call MUST carry traceparent")

	upstreamTC, ok3 := models.ParseTraceParent(upstreamTP)
	require.True(t, ok3)
	assert.Equal(t, incomingTraceID, upstreamTC.TraceID)
	assert.NotEqual(t, pdpTC.SpanID, upstreamTC.SpanID, "§3.2.2: each outgoing hop MUST get its own fresh child span")

	upstreamTS, hasTS := headerValue(ok.OkResponse.Headers, models.HeaderTraceState)
	require.True(t, hasTS, "§3.1: upstream call MUST also carry tracestate, not traceparent alone")
	assert.Equal(t, incomingTraceState, upstreamTS)

	upstreamFSC, hasFSC := headerValue(ok.OkResponse.Headers, models.HeaderFSCTransactionID)
	require.True(t, hasFSC)
	assert.Equal(t, fscTxnID, upstreamFSC)
}

// TestCheck_NoIncomingTraceContext_GeneratesFreshTraceSharedAcrossHops covers the "new trace" half of
// §3.2.2: absent an incoming traceparent, a fresh trace_id is minted once and shared by both hops rather
// than each hop generating its own unrelated trace.
func TestCheck_NoIncomingTraceContext_GeneratesFreshTraceSharedAcrossHops(t *testing.T) {
	t.Parallel()

	var pdpTraceParent string

	pdp := decisionPDP(t, true, func(r *http.Request) {
		pdpTraceParent = r.Header.Get(models.HeaderTraceParent)
	})
	defer pdp.Close()

	server := newTestAuthServer(t, pdp.URL)

	resp, err := server.Check(context.Background(), newCheckRequest(http.MethodGet, nil))
	require.NoError(t, err)

	pdpTC, ok := models.ParseTraceParent(pdpTraceParent)
	require.True(t, ok, "a fresh, valid traceparent MUST be generated even with no incoming header")

	ok2 := resp.HttpResponse.(*auth.CheckResponse_OkResponse)
	upstreamTP, _ := headerValue(ok2.OkResponse.Headers, models.HeaderTraceParent)
	upstreamTC, ok3 := models.ParseTraceParent(upstreamTP)
	require.True(t, ok3)

	assert.Equal(t, pdpTC.TraceID, upstreamTC.TraceID, "both hops MUST share the one freshly generated trace_id")
	assert.NotEqual(t, pdpTC.SpanID, upstreamTC.SpanID)
}

// TestCheck_TraceStateWithoutTraceParent_IsDropped covers §3.2.2's edge case (citing W3C §4.2-§4.3): a tracestate header
// arriving without a valid traceparent has nothing to attach to, so it MUST NOT be propagated as if it
// belonged to a freshly started trace.
func TestCheck_TraceStateWithoutTraceParent_IsDropped(t *testing.T) {
	t.Parallel()

	var (
		pdpTraceState    string
		pdpTraceStateSet bool
	)

	pdp := decisionPDP(t, true, func(r *http.Request) {
		pdpTraceState = r.Header.Get(models.HeaderTraceState)
		_, pdpTraceStateSet = r.Header["Tracestate"]
	})
	defer pdp.Close()

	server := newTestAuthServer(t, pdp.URL)

	req := newCheckRequest(http.MethodGet, map[string]string{
		models.HeaderTraceState: "vendorA=1",
	})

	_, err := server.Check(context.Background(), req)
	require.NoError(t, err)

	assert.False(t, pdpTraceStateSet, "tracestate MUST be dropped when there was no valid traceparent to carry it")
	assert.Empty(t, pdpTraceState)
}

// TestCheck_VersionedTraceParent_PreservesTraceState guards against a narrower reading of trace-context
// §3.2.4: a traceparent using a future version (here 01) still has a real, valid trace to attach to, so its
// tracestate MUST NOT be swept into the same "nothing to attach to" drop path as a missing/malformed header.
func TestCheck_VersionedTraceParent_PreservesTraceState(t *testing.T) {
	t.Parallel()

	const versionedTraceParent = "01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	var pdpTraceState string

	pdp := decisionPDP(t, true, func(r *http.Request) {
		pdpTraceState = r.Header.Get(models.HeaderTraceState)
	})
	defer pdp.Close()

	server := newTestAuthServer(t, pdp.URL)

	req := newCheckRequest(http.MethodGet, map[string]string{
		models.HeaderTraceParent: versionedTraceParent,
		models.HeaderTraceState:  "vendorA=1",
	})

	_, err := server.Check(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, "vendorA=1", pdpTraceState, "a valid non-00-version traceparent MUST NOT trigger the tracestate drop path")
}

// TestCheck_Denied_ReturnsDeniedResponse covers the deny path: a decision:false response from the PDP
// must short-circuit to a DeniedResponse rather than allowing the request through.
func TestCheck_Denied_ReturnsDeniedResponse(t *testing.T) {
	t.Parallel()

	pdp := decisionPDP(t, false, nil)
	defer pdp.Close()

	server := newTestAuthServer(t, pdp.URL)

	resp, err := server.Check(context.Background(), newCheckRequest(http.MethodGet, nil))
	require.NoError(t, err)

	denied, isDenied := resp.HttpResponse.(*auth.CheckResponse_DeniedResponse)
	require.True(t, isDenied, "expected a DeniedResponse for a decision:false PDP response")
	assert.Equal(t, types.StatusCode(http.StatusUnauthorized), denied.DeniedResponse.Status.Code)
}

// TestCheck_PDPUnreachable_ReturnsDeniedResponse covers the failure path: if the PDP can't be reached at
// all, the request must fail closed (denied), not be let through.
func TestCheck_PDPUnreachable_ReturnsDeniedResponse(t *testing.T) {
	t.Parallel()

	server := newTestAuthServer(t, "http://127.0.0.1:0")

	resp, err := server.Check(context.Background(), newCheckRequest(http.MethodGet, nil))
	require.NoError(t, err, "Check itself must not error out even when the PDP is unreachable")

	_, isDenied := resp.HttpResponse.(*auth.CheckResponse_DeniedResponse)
	assert.True(t, isDenied, "an unreachable PDP MUST fail closed")
}

// TestCheck_OPTIONS_BypassesPDPButStillPropagatesTrace covers the CORS preflight short-circuit: it must
// skip the PDP call entirely, yet still hand the upstream call a valid trace context (§3.1 applies
// regardless of which path allowed the request through).
func TestCheck_OPTIONS_BypassesPDPButStillPropagatesTrace(t *testing.T) {
	t.Parallel()

	pdpCalled := false

	pdp := decisionPDP(t, true, func(_ *http.Request) { pdpCalled = true })
	defer pdp.Close()

	server := newTestAuthServer(t, pdp.URL)

	const traceParent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	resp, err := server.Check(context.Background(), newCheckRequest(http.MethodOptions, map[string]string{
		models.HeaderTraceParent: traceParent,
	}))
	require.NoError(t, err)

	assert.False(t, pdpCalled, "OPTIONS MUST bypass the PDP call")

	ok, isOk := resp.HttpResponse.(*auth.CheckResponse_OkResponse)
	require.True(t, isOk)

	upstreamTP, hasTP := headerValue(ok.OkResponse.Headers, models.HeaderTraceParent)
	require.True(t, hasTP, "§3.1 still applies on the OPTIONS short-circuit path")

	upstreamTC, parsed := models.ParseTraceParent(upstreamTP)
	require.True(t, parsed)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", upstreamTC.TraceID)
}
