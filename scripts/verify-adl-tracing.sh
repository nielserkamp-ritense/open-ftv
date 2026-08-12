#!/usr/bin/env bash
#
# verify-adl-tracing.sh — verifies AuthZEN requests are persisted as ADL
# records via the manager API, for every request type that writes one:
# evaluation, batch evaluations, and all three search endpoints. For each,
# it also checks the W3C trace-context correlation fields (traceId, spanId,
# parentSpanId) and the resource (producer identification) field, not just
# eventName/status — this script's whole point is verifying tracing.
#
# The search endpoints are expected to come back with Status: Error — the
# vlierdam PIP test fixtures have no subjects/actions/resources enumeration
# data, which Search() (unlike Authorize/Batch) genuinely needs.
#
# A separate scenario sends an evaluation with no traceparent header at all,
# to verify the server-side fallback (a fresh trace/span pair, no parent)
# still produces a correlatable ADL entry.
#
# Run from the project root:
#   ./scripts/verify-adl-tracing.sh
#
# Environment:
#   KEEP_STACK=1   leave containers running after the script finishes
#   NO_BUILD=1     skip `docker compose build` (reuse existing images)
#
set -euo pipefail

COMPOSE="docker compose -f docker/compose.yaml"
SERVICES=(vlierdam-db vlierdam-dataspace vlierdam-manager vlierdam-pdp1)

MGR="http://localhost:9000"
PDP="http://localhost:9004"
PDP_HEALTH="http://localhost:8104"

RUN_ID="${RUN_ID:-$(date +%s)}"
SUBJECT='trace-known-'$RUN_ID

fails=0
EVAL_BODY=""
EVAL_CODE=""
RESP_CODE=""
TRACEPARENT=""
TRACE_HEX=""
SPAN_HEX=""

pass() { echo "ok  $*"; }
fail() { echo "FAIL $*"; fails=$((fails + 1)); }

# fresh_trace sets TRACEPARENT, TRACE_HEX, SPAN_HEX to a new random W3C
# trace-context, so each scenario's ADL entry can be polled independently.
fresh_trace() {
  eval "$(python3 - <<'PY'
import secrets

trace = secrets.token_hex(16)
span = secrets.token_hex(8)
print(f"TRACEPARENT='00-{trace}-{span}-01'")
print(f"TRACE_HEX='{trace}'")
print(f"SPAN_HEX='{span}'")
PY
)"
}

wait_http() { # url label [expected-code]
  local url="$1" label="$2" expect="${3:-200}" code
  echo ">> waiting for $label"
  for _ in $(seq 1 60); do
    code=$(curl -s -o /dev/null -w '%{http_code}' "$url" || true)
    if [[ "$code" == "$expect" ]]; then
      pass "$label ready ($code)"
      return 0
    fi
    sleep 3
  done
  fail "$label not ready (last status $code, expected $expect)"
  return 1
}

authzen_eval() { # traceparent subject_id; sets EVAL_BODY and EVAL_CODE
  local tp="$1" subject="$2"
  local headers=(-H 'Content-Type: application/json')
  [[ -n "$tp" ]] && headers+=(-H "traceparent: $tp")

  local tmp
  tmp=$(mktemp)
  EVAL_CODE=$(curl -s -o "$tmp" -w '%{http_code}' -X POST "${headers[@]}" \
    -d "$(cat <<EOF
{
  "subject": {"type": "medewerker", "id": "$subject"},
  "action": {"name": "GET"},
  "resource": {"type": "service", "id": "laadpalen"}
}
EOF
)" "$PDP/authzen/v1/evaluation")
  EVAL_BODY=$(cat "$tmp")
  rm -f "$tmp"
}

# post_authzen posts an arbitrary body to an arbitrary AuthZEN path, setting
# RESP_CODE. Used for the endpoints authzen_eval doesn't cover.
post_authzen() { # path traceparent json_body
  local path="$1" tp="$2" body="$3"
  local headers=(-H 'Content-Type: application/json')
  [[ -n "$tp" ]] && headers+=(-H "traceparent: $tp")

  local tmp
  tmp=$(mktemp)
  RESP_CODE=$(curl -s -o "$tmp" -w '%{http_code}' -X POST "${headers[@]}" -d "$body" "$PDP$path")
  rm -f "$tmp"
}

assert_decision() { # expected true|false
  local want="$1"
  EVAL_JSON="$EVAL_BODY" WANT="$want" python3 - <<'PY'
import json, os, sys
body = json.loads(os.environ["EVAL_JSON"])
want = os.environ["WANT"] == "true"
got = body.get("decision")
if got is not want:
    print(f"decision mismatch: want={want!r} got={got!r} body={body!r}")
    sys.exit(1)
print(f"decision={got!r}")
PY
}

wait_pdp_allows() {
  echo ">> waiting for PDP to permit evaluations"
  for _ in $(seq 1 40); do
    authzen_eval '' 'policy-probe-'$RUN_ID
    if [[ "$EVAL_CODE" == "200" ]] && assert_decision true; then
      pass "PDP returns decision=true"
      return 0
    fi
    sleep 3
  done
  fail "PDP never returned decision=true (last body: $EVAL_BODY)"
  return 1
}

poll_adl_trace() { # trace_hex -> sets ADL_BODY
  local trace="$1"
  ADL_BODY=""
  for _ in $(seq 1 20); do
    ADL_BODY=$(curl -sf "$MGR/v1/adl/entries?traceId=$trace&limit=5" 2>/dev/null || true)
    if [[ -n "$ADL_BODY" ]] && ADL_JSON="$ADL_BODY" python3 - <<'PY'
import json, os, sys
data = json.loads(os.environ["ADL_JSON"])
sys.exit(0 if isinstance(data, list) and len(data) > 0 else 1)
PY
    then
      return 0
    fi
    sleep 1
  done
  return 1
}

assert_adl_entry() {
  ADL_JSON="$ADL_BODY" TRACE_HEX="$TRACE_HEX" SPAN_HEX="$SPAN_HEX" SUBJECT="$SUBJECT" python3 - <<'PY'
import json, os, re, sys
entries = json.loads(os.environ["ADL_JSON"])
want_trace = os.environ["TRACE_HEX"]
want_parent = os.environ["SPAN_HEX"]
want_subject = os.environ["SUBJECT"]
match = next((e for e in entries if e.get("traceId") == want_trace), entries[0])
errors = []
if match.get("traceId") != want_trace:
    errors.append(f"traceId={match.get('traceId')!r}")
parent = match.get("parentSpanId") or ""
if parent != want_parent:
    errors.append(f"parentSpanId={parent!r}")
span = match.get("spanId") or ""
if not re.fullmatch(r"[0-9a-f]{16}", span):
    errors.append(f"spanId={span!r}")
if span == want_parent:
    errors.append("spanId must be a new child span, not the incoming traceparent span")
if match.get("eventName") != "adl.access_evaluation":
    errors.append(f"eventName={match.get('eventName')!r}")
if match.get("status") != "Ok":
    errors.append(f"status={match.get('status')!r} want='Ok'")
subj = ((match.get("request") or {}).get("subject") or {}).get("id")
if subj != want_subject:
    errors.append(f"subject.id={subj!r}")
resp = match.get("response") or {}
if resp.get("decision") is not True:
    errors.append(f"response.decision={resp.get('decision')!r}")
resource = match.get("resource") or {}
if not resource:
    errors.append("resource field is empty; expected producer identification (e.g. service name)")
if errors:
    print("ADL mismatch: " + ", ".join(errors))
    sys.exit(1)
print(
    f"ADL entry id={match.get('id')} trace={match.get('traceId')} "
    f"parentSpan={parent} span={span} decision={resp.get('decision')!r} resource={resource!r}"
)
PY
}

# assert_adl_event verifies an ADL entry exists for trace_hex with the given
# event_name and status, plus the same trace-context correlation fields
# (spanId/parentSpanId/resource) that assert_adl_entry checks for the plain
# evaluation case — every ADL record must carry these regardless of request
# type. When want_status is Error, also checks the response is empty (the
# PDP couldn't produce one).
assert_adl_event() { # trace_hex parent_span_hex want_event_name want_status
  local trace="$1" parent_hex="$2" want_event="$3" want_status="$4"
  if ! poll_adl_trace "$trace"; then
    fail "no ADL entry found for traceId=$trace within timeout"
    return 1
  fi

  ADL_JSON="$ADL_BODY" TRACE_HEX="$trace" PARENT_HEX="$parent_hex" WANT_EVENT="$want_event" WANT_STATUS="$want_status" python3 - <<'PY'
import json, os, re, sys
entries = json.loads(os.environ["ADL_JSON"])
want_trace = os.environ["TRACE_HEX"]
want_parent = os.environ["PARENT_HEX"]
want_event = os.environ["WANT_EVENT"]
want_status = os.environ["WANT_STATUS"]
match = next((e for e in entries if e.get("traceId") == want_trace), entries[0] if entries else None)
if match is None:
    print("no entries returned")
    sys.exit(1)
errors = []
if match.get("eventName") != want_event:
    errors.append(f"eventName={match.get('eventName')!r} want={want_event!r}")
if match.get("status") != want_status:
    errors.append(f"status={match.get('status')!r} want={want_status!r}")
if want_status == "Error" and match.get("response"):
    errors.append(f"response should be empty when status=Error, got {match.get('response')!r}")
if match.get("traceId") != want_trace:
    errors.append(f"traceId={match.get('traceId')!r} want={want_trace!r}")
parent = match.get("parentSpanId") or ""
if parent != want_parent:
    errors.append(f"parentSpanId={parent!r} want={want_parent!r}")
span = match.get("spanId") or ""
if not re.fullmatch(r"[0-9a-f]{16}", span):
    errors.append(f"spanId={span!r}")
if span == want_parent:
    errors.append("spanId must be a new child span, not the incoming traceparent span")
resource = match.get("resource") or {}
if not resource:
    errors.append("resource field is empty; expected producer identification (e.g. service name)")
if errors:
    print("ADL mismatch: " + ", ".join(errors))
    sys.exit(1)
print(
    f"ADL entry eventName={match.get('eventName')!r} status={match.get('status')!r} "
    f"span={span} parentSpan={parent!r} resource={resource!r}"
)
PY
}

# poll_adl_subject polls the manager for ADL entries by subjectId within a
# recent window, rather than by traceId — used when the trace/span pair is
# generated server-side and so isn't known ahead of the request.
poll_adl_subject() { # subject_id -> sets ADL_BODY
  local subject="$1"
  ADL_BODY=""
  for _ in $(seq 1 20); do
    ADL_BODY=$(curl -sf "$MGR/v1/adl/entries?subjectId=$subject&recent=5m&limit=5" 2>/dev/null || true)
    if [[ -n "$ADL_BODY" ]] && ADL_JSON="$ADL_BODY" python3 - <<'PY'
import json, os, sys
data = json.loads(os.environ["ADL_JSON"])
sys.exit(0 if isinstance(data, list) and len(data) > 0 else 1)
PY
    then
      return 0
    fi
    sleep 1
  done
  return 1
}

# assert_adl_fallback_trace verifies the ADL entry for subject_id has a
# server-generated trace/span pair: well-formed and correlatable, but with
# no parentSpanId, since there was no incoming traceparent to derive one
# from.
assert_adl_fallback_trace() { # subject_id
  ADL_JSON="$ADL_BODY" SUBJECT="$1" python3 - <<'PY'
import json, os, re, sys
entries = json.loads(os.environ["ADL_JSON"])
want_subject = os.environ["SUBJECT"]
match = next(
    (e for e in entries if ((e.get("request") or {}).get("subject") or {}).get("id") == want_subject),
    None,
)
if match is None:
    print(f"no entry found for subject.id={want_subject!r}")
    sys.exit(1)
errors = []
trace = match.get("traceId") or ""
if not re.fullmatch(r"[0-9a-f]{32}", trace):
    errors.append(f"traceId={trace!r} is not a valid 32-hex trace id")
span = match.get("spanId") or ""
if not re.fullmatch(r"[0-9a-f]{16}", span):
    errors.append(f"spanId={span!r} is not a valid 16-hex span id")
parent = match.get("parentSpanId") or ""
if parent != "":
    errors.append(f"parentSpanId={parent!r}, want empty (no incoming traceparent header)")
if match.get("eventName") != "adl.access_evaluation":
    errors.append(f"eventName={match.get('eventName')!r}")
if match.get("status") != "Ok":
    errors.append(f"status={match.get('status')!r} want='Ok'")
resource = match.get("resource") or {}
if not resource:
    errors.append("resource field is empty; expected producer identification (e.g. service name)")
if errors:
    print("ADL mismatch: " + ", ".join(errors))
    sys.exit(1)
print(f"ADL entry trace={trace} span={span} parentSpan={parent!r}")
PY
}

echo ">> bringing up tracing stack (${SERVICES[*]})"
if [[ "${NO_BUILD:-}" == "1" ]]; then
  $COMPOSE up -d "${SERVICES[@]}"
else
  $COMPOSE up -d --build "${SERVICES[@]}"
fi

wait_http "$MGR/v1/policies" "manager API"
wait_http "$PDP_HEALTH/livez" "PDP liveness"
wait_pdp_allows

echo ">> AuthZEN evaluation (authorization request)"
fresh_trace
authzen_eval "$TRACEPARENT" "$SUBJECT"
if [[ "$EVAL_CODE" == "200" ]]; then
  pass "evaluation request ($EVAL_CODE)"
else
  fail "evaluation request expected 200, got $EVAL_CODE"
fi
if assert_decision true; then
  pass "evaluation returns decision=true"
else
  fail "evaluation did not return decision=true"
fi

echo ">> ADL search by traceId"
if poll_adl_trace "$TRACE_HEX"; then
  if assert_adl_entry; then
    pass "ADL record persisted correctly for the authorization request"
  else
    fail "ADL entry fields mismatch"
  fi
else
  fail "no ADL entry found for traceId=$TRACE_HEX within timeout"
fi

echo ">> AuthZEN evaluation without a traceparent header (server-generated fallback)"
NO_TRACE_SUBJECT='trace-fallback-'$RUN_ID
authzen_eval '' "$NO_TRACE_SUBJECT"
if [[ "$EVAL_CODE" == "200" ]]; then
  pass "evaluation without traceparent header ($EVAL_CODE)"
else
  fail "evaluation without traceparent header expected 200, got $EVAL_CODE"
fi
if poll_adl_subject "$NO_TRACE_SUBJECT"; then
  if assert_adl_fallback_trace "$NO_TRACE_SUBJECT"; then
    pass "ADL record persisted with a server-generated trace/span (no parent)"
  else
    fail "ADL entry for header-less evaluation missing or mismatched"
  fi
else
  fail "no ADL entry found for subjectId=$NO_TRACE_SUBJECT within timeout"
fi

echo ">> AuthZEN batch evaluations"
fresh_trace
post_authzen "/authzen/v1/evaluations" "$TRACEPARENT" "$(cat <<EOF
{
  "evaluations": [
    {"subject": {"type": "medewerker", "id": "$SUBJECT"}, "action": {"name": "GET"}, "resource": {"type": "service", "id": "laadpalen"}}
  ]
}
EOF
)"
if [[ "$RESP_CODE" == "200" ]]; then
  pass "batch evaluations request ($RESP_CODE)"
else
  fail "batch evaluations expected 200, got $RESP_CODE"
fi
if assert_adl_event "$TRACE_HEX" "$SPAN_HEX" "adl.access_evaluations" "Ok"; then
  pass "ADL record persisted for batch evaluations"
else
  fail "ADL entry for batch evaluations missing or mismatched"
fi

echo ">> AuthZEN search: subject (no PIP enumeration data -> Status: Error)"
fresh_trace
post_authzen "/authzen/v1/search/subject" "$TRACEPARENT" "$(cat <<EOF
{
  "subject": {"type": "medewerker"},
  "action": {"name": "GET"},
  "resource": {"type": "service", "id": "laadpalen"}
}
EOF
)"
if [[ "$RESP_CODE" == "200" ]]; then
  pass "search subject request ($RESP_CODE)"
else
  fail "search subject expected 200, got $RESP_CODE"
fi
if assert_adl_event "$TRACE_HEX" "$SPAN_HEX" "adl.search_subject" "Error"; then
  pass "ADL record persisted for search subject (status=Error, response empty)"
else
  fail "ADL entry for search subject missing or mismatched"
fi

echo ">> AuthZEN search: action (no PIP enumeration data -> Status: Error)"
fresh_trace
post_authzen "/authzen/v1/search/action" "$TRACEPARENT" "$(cat <<EOF
{
  "subject": {"type": "medewerker", "id": "$SUBJECT"},
  "resource": {"type": "service", "id": "laadpalen"}
}
EOF
)"
if [[ "$RESP_CODE" == "200" ]]; then
  pass "search action request ($RESP_CODE)"
else
  fail "search action expected 200, got $RESP_CODE"
fi
if assert_adl_event "$TRACE_HEX" "$SPAN_HEX" "adl.search_action" "Error"; then
  pass "ADL record persisted for search action (status=Error, response empty)"
else
  fail "ADL entry for search action missing or mismatched"
fi

echo ">> AuthZEN search: resource (no PIP enumeration data -> Status: Error)"
fresh_trace
post_authzen "/authzen/v1/search/resource" "$TRACEPARENT" "$(cat <<EOF
{
  "subject": {"type": "medewerker", "id": "$SUBJECT"},
  "action": {"name": "GET"},
  "resource": {"type": "service"}
}
EOF
)"
if [[ "$RESP_CODE" == "200" ]]; then
  pass "search resource request ($RESP_CODE)"
else
  fail "search resource expected 200, got $RESP_CODE"
fi
if assert_adl_event "$TRACE_HEX" "$SPAN_HEX" "adl.search_resource" "Error"; then
  pass "ADL record persisted for search resource (status=Error, response empty)"
else
  fail "ADL entry for search resource missing or mismatched"
fi

echo
if [[ "$fails" -eq 0 ]]; then
  echo "PASS: ADL persistence verification succeeded"
else
  echo "FAIL: $fails assertion(s) failed"
  exit 1
fi

if [[ "${KEEP_STACK:-}" != "1" ]]; then
  echo ">> tearing down tracing stack"
  $COMPOSE down
else
  echo ">> KEEP_STACK=1 — containers left running"
fi
