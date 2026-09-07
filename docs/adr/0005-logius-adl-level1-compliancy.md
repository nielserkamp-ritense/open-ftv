# Logius Authorization Decision Log — Level 1 Conformance Statement

**Standard:** Logius Authorization Decision Log
**Scope:** Level 1
**Date:** 20 August 2026 (revised)

A plain statement of where the Policy Decision Point's decision log meets Level 1 of the Logius ADL standard, and where it doesn't yet. The § column cites the standard's own section numbers ([v1.0.0](https://gitdocumentatie.logius.nl/publicatie/ftv/adl/1.0.0/)); a row citing more than one means the requirement is assembled from several paragraphs, not one.

## Summary

| Status | Count |
|---|---:|
| ✅ Compliant | 19 |
| 🟡 Partially met | 0 |
| 🔴 Not yet met | 0 |
| 🔷 Depends on deployment | 1 |

Everything the standard requires — the record's contents, trace-context propagation across every component on the path, re-delivery safety, and writing it before the caller gets an answer — is met. The one remaining item is how the storage connection is secured, which is set per environment rather than by the application itself.

## The record

What Level 1 requires every decision record to contain, and whether it does.

| Requirement | § | Status | Notes |
|---|---|---|---|
| Trace and span identifiers | §3.3.1, §3.3.2 | ✅ Compliant | Every record carries a valid trace identifier and a valid span identifier, generated securely and never derived from anything user-identifiable, exactly as the standard requires. |
| Parent identifier | §3.3.3 | ✅ Compliant | Set from the caller's own identifier whenever a request arrives already part of a larger trace. Left out only when a request genuinely starts a new trace — exactly the one case the standard allows omitting it. The record's own span is now also linked to that parent through the tracing SDK's own mechanism, not just as a stored field, so tools reading the log over OTLP render the correct trace tree instead of seeing every record as a root span — an implementation-quality improvement in the spirit of §3.4's span/record correlation, though the standard itself only mandates the stored field. |
| Event name | §3.3.4, §3.4 | ✅ Compliant | Every record is labelled with one of the five event names the standard defines, matching the type of authorization request it belongs to. The span carrying the record is also named after it (§3.4, SHOULD), so tracing tools can identify authorization decisions without joining to the log. |
| Timestamp | §3.3.5 | ✅ Compliant | Records the moment the decision was actually made, not the moment the request arrived — a distinction the standard draws explicitly. |
| Status | §3.3.6 | ✅ Compliant | Every record is marked as succeeded or failed. A denied request is still a success in this sense — failure is reserved for the engine being unable to reach a decision at all, and a denial is never reported as a failure. |
| Request and response | §3.3.7.1, §3.3.7.2, §3.3.8 | ✅ Compliant | The full authorization request and response are carried in the record's body. When a response genuinely doesn't exist — an attempt that failed before one could be produced — it's left out of the record rather than filled in as an empty placeholder. |
| Source references | §3.3.7, §4.1.1 | ✅ Compliant | Left empty on every record, which is what Level 1 specifically calls for: the request and response are carried in full in the body rather than referenced elsewhere in attributes. |
| FSC TransactionID | §3.3.7.6 | ✅ Compliant | Set from the `Fsc-Transaction-Id` header when a request genuinely crosses an FSC inway or outway, and left out otherwise — the one case where the standard ties a field's presence to a fact about the request rather than to the detail level chosen. This no longer rides on whether a request/response body happens to be present; the two are tracked independently. |
| Producer identity | §3.3.9 | ✅ Compliant | Every record identifies which decision-point service produced it, so records collected from several deployments can still be traced back to their source. |

## Trace context propagation

What Level 1 requires of how trace context moves across every component on the path — not just the fields written into the record itself. This was previously undocumented; it's a distinct set of requirements from the record's own contents, covered by a different part of the standard (its behavioural rules for components, not the record's field list).

| Requirement | § | Status | Notes |
|---|---|---|---|
| Both trace-context headers used over HTTP | §3.1 | ✅ Compliant | Every hop that forwards a request — the gateway's call to the decision point, and the request it forwards on to the protected service — carries both `traceparent` and `tracestate`, not `traceparent` alone. |
| A fresh child span per outgoing call | §3.2.2 | ✅ Compliant | Each outgoing hop gets its own newly generated span identifier under the same trace, rather than reusing the identifier it received. The gateway's call to the decision point and its forwarded call to the protected service are treated as two distinct operations, each with its own identifier. |
| Trace identifier preserved, never restarted mid-flow | §3.2.2 | ✅ Compliant | A trace that's already underway keeps the same trace identifier all the way through; a new one is only ever generated when a request genuinely arrives without one. |
| Sampled flag left untouched | §3.2.2 | ✅ Compliant | The standard requires this flag to survive unmodified as it's propagated, and requires records to be produced regardless of its value. Both hold: the flag is carried through exactly as received, and record-writing doesn't consult it at all. |
| Vendor trace state dropped only when a new trace starts | §3.2.2 (citing trace-context §3.5 for the unmodified-passthrough rule, and §4.2-§4.3 for discarding it when there's no valid traceparent) | ✅ Compliant | Vendor-specific trace state is carried through unchanged whenever a trace continues, on both the call to the decision point and the call it forwards onward to the protected service — the upstream hop explicitly clears any stale tracestate rather than merely not re-setting it, closing a gap where a value belonging to a discarded trace could otherwise leak through unmodified. It's only ever dropped when the trace itself had to be restarted — never carried into a trace it doesn't belong to. Note: §3.5 and §4.2-§4.3 here are the W3C Trace Context standard's own sections, not this standard's own §3.5 (Sources and referencing) or §4 (Data Verifiability and Level of Detail), which are different topics. |

## Operational behaviour

What Level 1 requires of how records are produced and handled, beyond their contents.

| Requirement | § | Status | Notes |
|---|---|---|---|
| Honouring the caller's trace | §3.2.2, §3.3.3 | ✅ Compliant | An incoming request that already carries trace context is logged as part of that same trace, and that trace context now also propagates correctly to every component the request passes through — not just into the log record itself. When none is present, a new one is generated so the record still exists on its own. |
| Exactly one record per decision | §3.2.3 | ✅ Compliant | Every decision attempt — whether it completes, is denied, or fails outright — produces exactly one record. Nothing is dropped, and nothing is logged twice for a single request. |
| Written before answering the caller | §2.3.2 | ✅ Compliant | The standard recommends that a record reach durable storage before the caller is told the decision, so a crash immediately afterwards can't silently lose it. The decision path waits for the record to be written before returning, trading a small amount of latency per request for that guarantee. |
| Safe to re-deliver | §3.2.4 | ✅ Compliant | The standard requires that reprocessing the same record twice — for example after a network retry — must not create a duplicate. Records are uniquely keyed by their trace and span identifiers, so a repeat delivery is recognised and dropped instead of being logged twice. |
| Encrypted connection to the log | §3.2.1 | 🔷 Depends on deployment | The standard requires the connection to the log to be encrypted. That's set through the database connection configuration for each environment rather than enforced by the application itself, and it needs to be turned on explicitly — it is not the default in the example configurations we ship. |

## Scope note

> This statement covers what Level 1 requires the record to contain, the propagation of trace context across every component that participates in an authorization decision, and the core rules for producing and writing a record. It does not cover retention, access control to the log, or long-term integrity — the standard explicitly leaves those to each organisation's own policy rather than prescribing them.

## Non-normative considerations

The standard also marks §3.5 (source referencing), §4 (detail levels), and all of §5 (legal/security/access-control guidance) as non-normative — none of it gates Level 1 conformance, but worth a plain note on where things stand.

- §3.5 doesn't apply yet: at Level 1 nothing is referenced, everything is inlined in `body`.
- §4 is the level-selection framework itself, not a checklist; Level 1 was a deliberate choice, documented here.
- §5 is explicitly organisational (*"conformance to this standard does not, by itself, guarantee legal or regulatory compliance"*) — a written logging policy, DPIA, retention schedule, and Register of Processing Activities entry are outside what a codebase can satisfy on its own, and haven't been produced by anyone yet, in either direction.
- Two concrete gaps worth naming: the `dpl.core.data_subject_id` attribute (§5.1.4, SHOULD, tied to the separate LDV standard) isn't set anywhere; and nothing enforces append-only/WORM storage (§5.3) — the `decision` table can be updated or deleted like any other.
