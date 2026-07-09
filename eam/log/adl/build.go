package adl

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// InformationProvider is the hook through which the PIP supplies, per request, source
// references to the information it consulted (level 3, adl.core.information).
//
// The PIP-side implementation of this interface is provided by a separate component.
// The returned map is keyed by information-source name; each value MUST be a source
// reference (for example a Logged source {"span_id": "..."} or a version identifier),
// never a raw payload. Returning nil or an empty map records no information sources.
type InformationProvider interface {
	Information(ctx context.Context, ar *authlog.AuthRecord) map[string]any
}

// build maps an authlog.AuthRecord onto an ADL Record, honouring the configured level
// of detail. Exactly one Record is produced per call, for both successful and errored
// evaluations.
func (l *Logger) build(ctx context.Context, ar *authlog.AuthRecord) *Record {
	tc := deriveTraceContext(ar.TraceParent)

	rec := &Record{
		TraceID:      tc.traceID,
		SpanID:       tc.spanID,
		ParentSpanID: tc.parentSpanID,
		EventName:    eventName(ar),
		Timestamp:    timestamp(ar),
		Status:       status(ar),
		Resource:     l.cfg.resourceCopy(),
		Body:         map[string]any{},
	}

	// Level 1: request and response as raw payloads in body.
	rec.Body[KeyRequest] = authzenRequest(ar)
	if rec.Status != StatusError {
		// The response MUST be retrievable for Unset/Ok and MAY be omitted on Error.
		rec.Body[KeyResponse] = authzenResponse(ar)
	}

	attrs := map[string]any{}

	// F18: adl.fsc.transaction_id MUST be set when the decision request crossed an FSC
	// inway/outway - the FSC handler extracted the Fsc-Transaction-Id header - and MUST
	// NOT be set otherwise (a pure AuthZEN PDP request carries no FSC transaction). Only
	// the FSC path populates ar.FSCTransactionID, so this is empty (and omitted) elsewhere.
	if ar.FSCTransactionID != "" {
		attrs[KeyTransactionID] = ar.FSCTransactionID
	}

	// Level 2: reference the exact version of the policies used (source reference).
	//
	// Level 2 is only reached WITH a resolvable policy hash: {policyKey: policyHash} is a
	// version reference from which the exact policy text can be retrieved. When the engine
	// supplied a policy key but no hash we deliberately OMIT adl.core.policies rather than
	// emit a {key: key} pseudo-reference, which is not resolvable to a policy version and
	// would falsely present the record as "level 2 satisfied". The record is degraded
	// explicitly (attribute absent + a warning) so it never claims an unresolvable version.
	if l.cfg.Level >= 2 {
		if pol := policiesRef(ar); len(pol) > 0 {
			attrs[KeyPolicies] = pol
		} else if key := decisionString(ar, "policy"); key != "" && l.cfg.Logger != nil {
			l.cfg.Logger.Warn("adl: omitting adl.core.policies: no resolvable policy hash for this decision",
				"policy", key, "trace_id", tc.traceID, "span_id", tc.spanID)
		}
	}

	// Level 3: reference the information sources for this request.
	//
	// Honest limitation: the provider reports every registered source reference whose
	// key matches something the request delivered (the entity UIDs, their attributes
	// keyed "<uid>/<attr>", and the context attribute keys). It does NOT prove which
	// of those the policy engine actually read during evaluation - full per-request
	// consultation tracking would require instrumenting the engine's attribute reads
	// and threading a per-request trace through it. This over-reports (references a
	// value that was available but perhaps unused) rather than under-reports, so it
	// never hides a consulted source; tightening it to exactly-consulted remains a
	// known, documented limitation.
	if l.cfg.Level >= 3 && l.info != nil {
		if info := l.info.Information(ctx, ar); len(info) > 0 {
			attrs[KeyInformation] = info
		}
	}

	// Level 4: capture the engine/configuration needed to reconstruct the environment.
	// These are raw key/value pairs describing the engine, so per the standard they
	// belong in body (not as an attributes source reference).
	if l.cfg.Level >= 4 {
		if conf := l.cfg.configurationBody(); len(conf) > 0 {
			rec.Body[KeyConfiguration] = conf
		}
	}

	// IM2 (SHOULD): promote a data-subject reference the application supplied on the
	// request context into the record attributes. This is a source-style reference
	// (which data subject the decision concerns, for DSAR / LDV linkage), not payload,
	// so it belongs in attributes. OpenFTV does not implement the LDV itself; it only
	// carries the reference through when the application delivered it.
	if ar.RequestContext != nil {
		if id := toString(ar.RequestContext.GetAttributeValue(KeyDataSubjectID)); id != "" {
			attrs[KeyDataSubjectID] = id
			if typ := toString(ar.RequestContext.GetAttributeValue(KeyDataSubjectType)); typ != "" {
				attrs[KeyDataSubjectType] = typ
			}
		}
	}

	if len(attrs) > 0 {
		rec.Attributes = attrs
	}

	return rec
}

// eventName resolves the ADL event name for the record, defaulting to the Access
// Evaluation API when the handler did not set one explicitly.
func eventName(ar *authlog.AuthRecord) string {
	if ar.EventName != "" {
		return ar.EventName
	}
	return EventAccessEvaluation
}

// status maps the evaluation outcome onto an ADL status. A completed evaluation - permit
// or deny - is Ok; only a PDP that could not evaluate yields Error.
func status(ar *authlog.AuthRecord) Status {
	if ar.Errored {
		return StatusError
	}
	return StatusOk
}

func timestamp(ar *authlog.AuthRecord) uint64 {
	t := time.Now().UTC()
	if ar.RequestTime != nil && !ar.RequestTime.IsZero() {
		t = *ar.RequestTime
	}
	return uint64(t.UnixMilli())
}

// authzenRequest reconstructs the full AuthZEN request from the record's PARC entities.
func authzenRequest(ar *authlog.AuthRecord) map[string]any {
	req := map[string]any{}

	if ar.Principal != nil {
		subject := map[string]any{"type": ar.Principal.Type(), "id": ar.Principal.ID()}
		if props := models.MapFromAttributes(ar.Principal.Attributes()); len(props) > 0 {
			subject["properties"] = props
		}
		req["subject"] = subject
	}

	if ar.Action != nil {
		action := map[string]any{"name": ar.Action.ID()}
		if props := models.MapFromAttributes(ar.Action.Attributes()); len(props) > 0 {
			action["properties"] = props
		}
		req["action"] = action
	}

	if ar.Resource != nil {
		resource := map[string]any{"type": ar.Resource.Type(), "id": ar.Resource.ID()}
		if props := models.MapFromAttributes(ar.Resource.Attributes()); len(props) > 0 {
			resource["properties"] = props
		}
		req["resource"] = resource
	}

	if ctx := models.MapFromAttributes(ar.RequestContext); len(ctx) > 0 {
		req["context"] = ctx
	}

	return req
}

// authzenResponse renders the AuthZEN response: the decision plus the response
// context (reason and any obligations). The obligations are the "ja, mits"
// duties the ODRL/ODRL-geo engine attached to a permit (auth_process.go stores
// them in DecisionContext under "obligations"); they belong in the logged
// response context so the ADL record faithfully reproduces what the PDP returned
// to the PEP, not just the bare decision.
func authzenResponse(ar *authlog.AuthRecord) map[string]any {
	resp := map[string]any{"decision": ar.Decision}
	ctx := map[string]any{}
	if msg := decisionString(ar, "message"); msg != "" {
		ctx["reason"] = map[string]any{"0": msg}
	}
	if obs := obligationsFrom(ar); len(obs) > 0 {
		ctx["obligations"] = obs
	}
	if len(ctx) > 0 {
		resp["context"] = ctx
	}
	return resp
}

// obligationsFrom extracts the AuthZEN obligations the handler placed in the
// decision context (a []map[string]any, mirroring the wire shape). Returns nil
// when the decision carried no obligations.
func obligationsFrom(ar *authlog.AuthRecord) []map[string]any {
	if ar.DecisionContext == nil {
		return nil
	}
	switch o := ar.DecisionContext.GetAttributeValue("obligations").(type) {
	case []map[string]any:
		return o
	case []any:
		out := make([]map[string]any, 0, len(o))
		for _, e := range o {
			if m, ok := e.(map[string]any); ok {
				out = append(out, m)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}
	return nil
}

// policiesRef builds the adl.core.policies source reference from the policy key and hash
// present in the decision context: {"<policyKey>": "<policyHash>"}. It returns nil unless
// BOTH a key and a resolvable content hash are available: the value MUST be a reference
// from which the exact policy version can be retrieved. A key repeated as its own value
// ({key: key}) is not resolvable, so it is never emitted - the caller degrades the record
// instead of presenting a non-resolvable version reference.
func policiesRef(ar *authlog.AuthRecord) map[string]any {
	key := decisionString(ar, "policy")
	if key == "" {
		return nil
	}
	hash := decisionString(ar, "policyHash")
	if hash == "" {
		return nil
	}
	return map[string]any{key: hash}
}

func decisionString(ar *authlog.AuthRecord, key string) string {
	if ar.DecisionContext == nil {
		return ""
	}
	return toString(ar.DecisionContext.GetAttributeValue(key))
}

func toString(v any) string {
	switch s := v.(type) {
	case nil:
		return ""
	case string:
		return s
	default:
		return fmt.Sprint(v)
	}
}
