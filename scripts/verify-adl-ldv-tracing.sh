#!/usr/bin/env bash
#
# verify-adl-ldv-tracing.sh — end-to-end ADL/LDV trace correlation checks.
#
# Brings up the minimal gemeente-vlierdam stack (manager + PDP + PostgreSQL ADL),
# sends AuthZEN evaluations with and without a W3C traceparent header, then
# verifies that:
#   1. PDP evaluations return decision=true (bootstrap permit-all policy)
#   2. ADL entries store the expected trace/span identifiers
#   3. LDV lezen returns ProcessingActivities mapped from those ADL rows
#   4. Auto-generated trace IDs are persisted when no traceparent is sent
#
# Run from the project root:
#   ./scripts/verify-adl-ldv-tracing.sh
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
SUBJECT_KNOWN='trace-known-'$RUN_ID
SUBJECT_AUTO='trace-auto-'$RUN_ID

# Fresh W3C traceparent per run (avoids stale ADL matches from prior runs).
eval "$(python3 - <<'PY'
import secrets

trace = secrets.token_hex(16)
span = secrets.token_hex(8)
print(f"TRACEPARENT='00-{trace}-{span}-01'")
print(f"TRACE_HEX='{trace}'")
print(f"SPAN_HEX='{span}'")
print(
    "TRACE_UUID='"
    f"{trace[0:8]}-{trace[8:12]}-{trace[12:16]}-{trace[16:20]}-{trace[20:32]}'"
)
PY
)"

fails=0
EVAL_BODY=""
EVAL_CODE=""

pass() { echo "ok  $*"; }
fail() { echo "FAIL $*"; fails=$((fails + 1)); }

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

poll_adl_subject() { # subject_id -> sets ADL_BODY
  local subject="$1"
  ADL_BODY=""
  for _ in $(seq 1 20); do
    ADL_BODY=$(curl -sf "$MGR/v1/adl/entries?subjectId=$subject&recent=5m&limit=10" 2>/dev/null || true)
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

assert_adl_known() {
  ADL_JSON="$ADL_BODY" TRACE_HEX="$TRACE_HEX" SPAN_HEX="$SPAN_HEX" SUBJECT="$SUBJECT_KNOWN" python3 - <<'PY'
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
if match.get("status") not in ("Unset", "Ok", "Error"):
    errors.append(f"status={match.get('status')!r}")
subj = ((match.get("request") or {}).get("subject") or {}).get("id")
if subj != want_subject:
    errors.append(f"subject.id={subj!r}")
resp = match.get("response") or {}
if resp.get("decision") is not True:
    errors.append(f"response.decision={resp.get('decision')!r}")
if errors:
    print("ADL mismatch: " + ", ".join(errors))
    sys.exit(1)
print(
    f"ADL entry id={match.get('id')} trace={match.get('traceId')} "
    f"parentSpan={parent} span={span} decision={resp.get('decision')!r}"
)
PY
}

assert_ldv_mapping() {
  TRACE_UUID="$TRACE_UUID" SPAN_HEX="$SPAN_HEX" SUBJECT="$SUBJECT_KNOWN" python3 - "$1" <<'PY'
import json, os, re, sys
path = sys.argv[1]
with open(path) as f:
    data = json.load(f)
want_uuid = os.environ["TRACE_UUID"]
want_parent = os.environ["SPAN_HEX"]
want_subject = os.environ["SUBJECT"]
acts = data.get("processingActivities") or []
if not acts:
    print("no processingActivities returned")
    sys.exit(1)
act = next((a for a in acts if a.get("traceId") == want_uuid), acts[0])
errors = []
if act.get("traceId") != want_uuid:
    errors.append(f"traceId={act.get('traceId')!r}")
if act.get("parentSpanId") != want_parent:
    errors.append(f"parentSpanId={act.get('parentSpanId')!r}")
span = act.get("spanId") or ""
if not re.fullmatch(r"[0-9a-f]{16}", span):
    errors.append(f"spanId={span!r}")
if act.get("name") != "GET":
    errors.append(f"name={act.get('name')!r}")
attrs = (act.get("resource") or {}).get("attributes") or {}
subj = (attrs.get("subject") or {}).get("id")
if subj != want_subject:
    errors.append(f"subject.id={subj!r}")
if errors:
    print("LDV mapping mismatch: " + ", ".join(errors))
    sys.exit(1)
print(f"LDV activity trace={act.get('traceId')} name={act.get('name')} status={act.get('status')}")
PY
}

assert_adl_auto() {
  ADL_JSON="$ADL_BODY" SUBJECT="$SUBJECT_AUTO" python3 - <<'PY'
import json, os, re, sys
entries = json.loads(os.environ["ADL_JSON"])
want_subject = os.environ["SUBJECT"]
match = next((e for e in entries if (e.get("request") or {}).get("subject", {}).get("id") == want_subject), None)
if not match:
    print(f"no entry for subject {want_subject!r}")
    sys.exit(1)
trace = match.get("traceId") or ""
span = match.get("spanId") or ""
errors = []
if not re.fullmatch(r"[0-9a-f]{32}", trace):
    errors.append(f"traceId={trace!r}")
if not re.fullmatch(r"[0-9a-f]{16}", span):
    errors.append(f"spanId={span!r}")
if match.get("parentSpanId"):
    errors.append(f"parentSpanId should be empty without traceparent, got {match.get('parentSpanId')!r}")
resp = match.get("response") or {}
if resp.get("decision") is not True:
    errors.append(f"response.decision={resp.get('decision')!r}")
if errors:
    print("auto trace invalid: " + ", ".join(errors))
    sys.exit(1)
print(f"auto trace traceId={trace} spanId={span} decision={resp.get('decision')!r}")
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

echo ">> AuthZEN evaluation with known traceparent"
authzen_eval "$TRACEPARENT" "$SUBJECT_KNOWN"
if [[ "$EVAL_CODE" == "200" ]]; then
  pass "evaluation with traceparent ($EVAL_CODE)"
else
  fail "evaluation with traceparent expected 200, got $EVAL_CODE"
fi
if assert_decision true; then
  pass "evaluation with traceparent returns decision=true"
else
  fail "evaluation with traceparent did not return decision=true"
fi

echo ">> ADL search by traceId"
if poll_adl_trace "$TRACE_HEX"; then
  if assert_adl_known; then
    pass "ADL stores traceparent correlation"
  else
    fail "ADL entry fields mismatch for known traceparent"
  fi
else
  fail "no ADL entry found for traceId=$TRACE_HEX within timeout"
fi

echo ">> LDV lezen by trace UUID"
ldv_tmp=$(mktemp)
ldv_headers=$(mktemp)
ldv_code=$(curl -s -D "$ldv_headers" -o "$ldv_tmp" -w '%{http_code}' -X POST \
  -H 'Content-Type: application/json' \
  -d "{\"traceId\":\"$TRACE_UUID\"}" \
  "$MGR/v1/lezenapi/v1/processing-activities")
ldv_api_version=$(awk -F': ' 'tolower($1)=="api-version"{print $2}' "$ldv_headers" | tr -d '\r')

if [[ "$ldv_code" == "200" ]]; then
  pass "LDV lezen HTTP 200"
else
  fail "LDV lezen expected 200, got $ldv_code"
fi

if [[ "$ldv_api_version" == "1.1.0" ]]; then
  pass "LDV lezen API-Version header ($ldv_api_version)"
else
  fail "LDV lezen API-Version expected 1.1.0, got ${ldv_api_version:-<missing>}"
fi

if assert_ldv_mapping "$ldv_tmp"; then
  pass "LDV lezen maps ADL entry to ProcessingActivity"
else
  fail "LDV lezen response does not match ADL data"
fi
rm -f "$ldv_tmp" "$ldv_headers"

echo ">> AuthZEN evaluation without traceparent (auto-generated trace)"
authzen_eval '' "$SUBJECT_AUTO"
if [[ "$EVAL_CODE" == "200" ]]; then
  pass "evaluation without traceparent ($EVAL_CODE)"
else
  fail "evaluation without traceparent expected 200, got $EVAL_CODE"
fi
if assert_decision true; then
  pass "evaluation without traceparent returns decision=true"
else
  fail "evaluation without traceparent did not return decision=true"
fi

echo ">> ADL search for auto-generated trace"
if poll_adl_subject "$SUBJECT_AUTO"; then
  if assert_adl_auto; then
    pass "ADL stores auto-generated trace identifiers"
  else
    fail "auto-generated trace/span not valid in ADL"
  fi
else
  fail "no ADL entry found for subject=$SUBJECT_AUTO within timeout"
fi

echo
if [[ "$fails" -eq 0 ]]; then
  echo "PASS: ADL/LDV tracing verification succeeded"
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
