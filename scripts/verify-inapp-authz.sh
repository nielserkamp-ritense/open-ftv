#!/usr/bin/env bash
#
# verify-inapp-authz.sh — e2e role matrix for the APP-ONLY management plane.
# The manager itself validates the OIDC JWT and authorizes via its embedded Cedar PDP
# (no gateway). Brings up docker/compose-mgmt-inapp.yaml, gets a token per role from
# Keycloak, calls the manager API directly, and asserts the role matrix.
#
# Run from the project root:  ./scripts/verify-inapp-authz.sh
#
set -euo pipefail

COMPOSE="docker compose -f docker/compose-mgmt-inapp.yaml"
KC="http://localhost:8088"
MGR="http://localhost:9100"

echo ">> bringing up the app-only stack (manager = PEP)"
$COMPOSE up -d --build

echo ">> waiting for the Keycloak realm"
until curl -sf "$KC/realms/openftv/.well-known/openid-configuration" >/dev/null; do sleep 4; done

# The manager fetches the JWKS at startup; if it started before the realm was imported,
# restart it once so it (re)builds the keyfunc against a ready Keycloak.
echo ">> (re)starting the manager so it picks up the JWKS"
$COMPOSE restart manager >/dev/null

echo ">> waiting for the manager API"
until [ "$(curl -s -o /dev/null -w '%{http_code}' "$MGR/v1/policies")" != "000" ]; do sleep 3; done

token() { # role -> access token
  curl -s -d grant_type=password -d client_id=openftv -d "username=$1-user" -d password=password \
    "$KC/realms/openftv/protocol/openid-connect/token" \
    | python3 -c 'import sys,json; print(json.load(sys.stdin).get("access_token",""))'
}

code() { # role method path -> HTTP status from the manager
  local t; t=$(token "$1")
  curl -s -o /dev/null -w '%{http_code}' -X "$2" -H "Authorization: Bearer $t" "$MGR$3"
}

fails=0
is_deny() { [ "$1" = 401 ] || [ "$1" = 403 ]; }
allow() { if is_deny "$1"; then echo "FAIL expected allow, got $1 ($2)"; fails=$((fails+1)); else echo "ok  allow ($1) $2"; fi; }
deny()  { if is_deny "$1"; then echo "ok  deny  ($1) $2"; else echo "FAIL expected deny, got $1 ($2)"; fails=$((fails+1)); fi; }

echo ">> role matrix (manager enforces in-process)"
allow "$(code auditor GET  /v1/policies)"   "auditor read policies"
deny  "$(code auditor POST /v1/policy/x)"   "auditor write policy"
allow "$(code author  GET  /v1/policies)"   "author read policies"
allow "$(code author  POST /v1/policy/x)"   "author write policy"
deny  "$(code author  POST /v1/deployment)" "author publish (admin-only)"
allow "$(code admin   POST /v1/deployment)" "admin publish"

NOTOK=$(curl -s -o /dev/null -w '%{http_code}' "$MGR/v1/policies")
if is_deny "$NOTOK"; then echo "ok  no-token rejected ($NOTOK)"; else echo "FAIL no-token got $NOTOK"; fails=$((fails+1)); fi

echo
if [ "$fails" -eq 0 ]; then
  echo "PASS: in-app role matrix passed (the manager is the PEP)"
else
  echo "FAIL: $fails assertion(s) failed"; exit 1
fi
