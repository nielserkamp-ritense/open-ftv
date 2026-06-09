#!/usr/bin/env bash
set -euo pipefail

COMPOSE="docker compose -f docker/compose-mgmt-authz.yaml"
KC="http://localhost:8088"
GW="http://localhost:9200"

echo ">> bringing up the management-authz stack"
$COMPOSE up -d --build

echo ">> waiting for Keycloak realm"
until curl -sf "$KC/realms/openftv/.well-known/openid-configuration" >/dev/null; do sleep 4; done
echo ">> waiting for gateway"
until [ -n "$(curl -s -o /dev/null -w '%{http_code}' "$GW/v1/policies")" ] && [ "$(curl -s -o /dev/null -w '%{http_code}' "$GW/v1/policies")" != "000" ]; do sleep 3; done

token() {
  # The gateway plugin validates the token "iss" against the *internal* Keycloak
  # URL (http://keycloak:8080/realms/openftv). Keycloak derives the issuer from the
  # request Host header, so we override it here to match what the plugin enforces —
  # otherwise every token is rejected and the PDP sees an "invalid" principal.
  curl -s -H 'Host: keycloak:8080' \
    -d grant_type=password -d client_id=openftv -d "username=$1-user" -d password=password \
    "$KC/realms/openftv/protocol/openid-connect/token" \
    | python3 -c 'import sys,json; print(json.load(sys.stdin).get("access_token",""))'
}

code() { # role method path
  local t; t=$(token "$1")
  curl -s -o /dev/null -w '%{http_code}' -X "$2" -H "Authorization: Bearer $t" "$GW$3"
}

fails=0
allow() { # got label  -> pass unless 401/403
  if [ "$1" = 401 ] || [ "$1" = 403 ]; then echo "FAIL expected allow, got $1 ($2)"; fails=$((fails+1)); else echo "ok  allow ($1) $2"; fi; }
deny() { # got label -> pass only if 403
  if [ "$1" = 403 ]; then echo "ok  deny  (403) $2"; else echo "FAIL expected 403, got $1 ($2)"; fails=$((fails+1)); fi; }

allow "$(code auditor GET  /v1/policies)"   "auditor read policies"
deny  "$(code auditor POST /v1/policy/x)"   "auditor write policy"
allow "$(code author  GET  /v1/policies)"   "author read policies"
allow "$(code author  POST /v1/policy/x)"   "author write policy"
deny  "$(code author  POST /v1/deployment)" "author publish (admin-only)"
allow "$(code admin   POST /v1/deployment)" "admin publish"

NOTOK=$(curl -s -o /dev/null -w '%{http_code}' "$GW/v1/policies")
if [ "$NOTOK" = 401 ] || [ "$NOTOK" = 403 ]; then echo "ok  no-token rejected ($NOTOK)"; else echo "FAIL no-token got $NOTOK"; fails=$((fails+1)); fi

echo
if [ "$fails" -eq 0 ]; then echo "PASS: role-matrix passed"; else echo "FAIL: $fails assertion(s) failed"; exit 1; fi
