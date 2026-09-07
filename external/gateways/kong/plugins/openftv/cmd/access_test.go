package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kong/go-pdk/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Verifies trace context is propagated correctly across the two hops Kong sits on: the call to the PDP,
// and the call it forwards onward to the upstream service.

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

// TestAccess_Allowed_PropagatesTraceContextAcrossBothHops verifies §3.1 (both trace-context headers sent
// on every hop) and §3.2.2 (each hop gets its own fresh child span under the same trace_id, and
// tracestate is carried through unchanged).
func TestAccess_Allowed_PropagatesTraceContextAcrossBothHops(t *testing.T) {
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

	env, err := test.New(t, test.Request{
		Method: "GET",
		Url:    "https://gw.test/resource",
		Headers: http.Header{
			"Traceparent":        {incomingTraceParent},
			"Tracestate":         {incomingTraceState},
			"Fsc-Transaction-Id": {fscTxnID},
		},
	})
	require.NoError(t, err)

	env.DoAccess(&Config{PDP: pdp.URL, Timeout: 5000})

	// §3.1: both trace-context headers, on the hop to the PDP...
	pdpTC, ok := models.ParseTraceParent(pdpTraceParent)
	require.True(t, ok, "PDP call must carry a valid traceparent")
	assert.Equal(t, incomingTraceID, pdpTC.TraceID, "§3.2.2: trace_id MUST be preserved")
	assert.NotEqual(t, incomingSpanID, pdpTC.SpanID, "§3.2.2: the PDP hop MUST mint its own child span")
	assert.Equal(t, incomingTraceState, pdpTraceState, "§3.2.2/W3C §3.5: tracestate MUST be carried through unchanged")
	assert.Equal(t, fscTxnID, pdpFSCTxnID)

	// ...and on the forwarded hop to the upstream service.
	upstreamTP := env.ServiceReq.Headers.Get(models.HeaderTraceParent)
	upstreamTC, ok2 := models.ParseTraceParent(upstreamTP)
	require.True(t, ok2, "§3.1: upstream call MUST carry a valid traceparent")
	assert.Equal(t, incomingTraceID, upstreamTC.TraceID)
	assert.NotEqual(t, pdpTC.SpanID, upstreamTC.SpanID, "§3.2.2: each outgoing hop MUST get its own fresh child span")

	assert.Equal(t, incomingTraceState, env.ServiceReq.Headers.Get(models.HeaderTraceState),
		"§3.1: upstream call MUST also carry tracestate, not traceparent alone")
	assert.Equal(t, fscTxnID, env.ServiceReq.Headers.Get(models.HeaderFSCTransactionID))
}

// TestAccess_NoIncomingTraceContext_GeneratesFreshTraceSharedAcrossHops covers the "new trace" half of
// §3.2.2: absent an incoming traceparent, a fresh trace_id is minted once and shared by both hops rather
// than each hop generating its own unrelated trace.
func TestAccess_NoIncomingTraceContext_GeneratesFreshTraceSharedAcrossHops(t *testing.T) {
	var pdpTraceParent string

	pdp := decisionPDP(t, true, func(r *http.Request) {
		pdpTraceParent = r.Header.Get(models.HeaderTraceParent)
	})
	defer pdp.Close()

	env, err := test.New(t, test.Request{Method: "GET", Url: "https://gw.test/resource"})
	require.NoError(t, err)

	env.DoAccess(&Config{PDP: pdp.URL, Timeout: 5000})

	pdpTC, ok := models.ParseTraceParent(pdpTraceParent)
	require.True(t, ok, "a fresh, valid traceparent MUST be generated even with no incoming header")

	upstreamTC, ok2 := models.ParseTraceParent(env.ServiceReq.Headers.Get(models.HeaderTraceParent))
	require.True(t, ok2)

	assert.Equal(t, pdpTC.TraceID, upstreamTC.TraceID, "both hops MUST share the one freshly generated trace_id")
	assert.NotEqual(t, pdpTC.SpanID, upstreamTC.SpanID)
}

// TestAccess_TraceStateWithoutTraceParent_IsDropped covers §3.2.2's edge case (citing W3C §4.2-§4.3): a tracestate header
// arriving without a valid traceparent has nothing to attach to, so it MUST NOT be propagated as if it
// belonged to a freshly started trace.
func TestAccess_TraceStateWithoutTraceParent_IsDropped(t *testing.T) {
	var pdpTraceStateSet bool

	pdp := decisionPDP(t, true, func(r *http.Request) {
		_, pdpTraceStateSet = r.Header["Tracestate"]
	})
	defer pdp.Close()

	env, err := test.New(t, test.Request{
		Method:  "GET",
		Url:     "https://gw.test/resource",
		Headers: http.Header{"Tracestate": {"vendorA=1"}},
	})
	require.NoError(t, err)

	env.DoAccess(&Config{PDP: pdp.URL, Timeout: 5000})

	assert.False(t, pdpTraceStateSet, "tracestate MUST be dropped when there was no valid traceparent to carry it")

	_, upstreamHasTS := env.ServiceReq.Headers["Tracestate"]
	assert.False(t, upstreamHasTS,
		"the orphaned tracestate MUST be cleared on the upstream leg too, not just left unset for the PDP call")
}

// TestAccess_VersionedTraceParent_PreservesTraceState guards against a narrower reading of trace-context
// §3.2.4: a traceparent using a future version (here 01) still has a real, valid trace to attach to, so its
// tracestate MUST NOT be swept into the same "nothing to attach to" drop path as a missing/malformed header.
func TestAccess_VersionedTraceParent_PreservesTraceState(t *testing.T) {
	const versionedTraceParent = "01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	var pdpTraceState string

	pdp := decisionPDP(t, true, func(r *http.Request) {
		pdpTraceState = r.Header.Get(models.HeaderTraceState)
	})
	defer pdp.Close()

	env, err := test.New(t, test.Request{
		Method: "GET",
		Url:    "https://gw.test/resource",
		Headers: http.Header{
			"Traceparent": {versionedTraceParent},
			"Tracestate":  {"vendorA=1"},
		},
	})
	require.NoError(t, err)

	env.DoAccess(&Config{PDP: pdp.URL, Timeout: 5000})

	assert.Equal(t, "vendorA=1", pdpTraceState, "a valid non-00-version traceparent MUST NOT trigger the tracestate drop path")
	assert.Equal(t, "vendorA=1", env.ServiceReq.Headers.Get(models.HeaderTraceState))
}

// TestAccess_Denied_ExitsForbiddenWithoutForwarding covers the deny path: a decision:false response
// from the PDP must exit with 403 and must not apply any upstream trace headers.
func TestAccess_Denied_ExitsForbiddenWithoutForwarding(t *testing.T) {
	pdp := decisionPDP(t, false, nil)
	defer pdp.Close()

	env, err := test.New(t, test.Request{
		Method:  "GET",
		Url:     "https://gw.test/resource",
		Headers: http.Header{"Traceparent": {"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"}},
	})
	require.NoError(t, err)

	env.DoAccess(&Config{PDP: pdp.URL, Timeout: 5000})

	assert.Equal(t, http.StatusForbidden, env.ClientRes.Status)
}

// TestAccess_PDPUnreachable_ExitsInternalServerError covers the failure path: if the PDP can't be
// reached at all, the request must fail closed with a 500, not be let through.
func TestAccess_PDPUnreachable_ExitsInternalServerError(t *testing.T) {
	// A closed listener gives an immediate, OS-level connection-refused error, unlike dialing port 0
	// (whose ephemeral-port allocation path is slower and more prone to perturbing goroutine scheduling
	// in the go-pdk test bridge's background reader).
	closedPDP := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	closedPDP.Close()

	env, err := test.New(t, test.Request{Method: "GET", Url: "https://gw.test/resource"})
	require.NoError(t, err)

	env.DoAccess(&Config{PDP: closedPDP.URL, Timeout: 5000})

	assert.Equal(t, http.StatusInternalServerError, env.ClientRes.Status)
}
