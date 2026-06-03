#!/usr/bin/env bash
#
# scripts/test.sh — draait alle tests (unit + e2e) in een container.
# Niks lokaal nodig behalve Docker; er hoeft geen Go geïnstalleerd te zijn.
#
# Werking: dit script start een golang-container en voert zichzelf daarbinnen
# opnieuw uit (self-re-exec). De module-/build-cache leeft in docker-volumes,
# dus de eerste run downloadt dependencies en latere runs zijn snel.
#
# Gebruik:
#   ./scripts/test.sh              # alle unit-tests + e2e
#   ./scripts/test.sh --unit       # alleen unit-tests
#   ./scripts/test.sh --e2e        # alleen e2e
#   ./scripts/test.sh apps/pdp ... # alleen deze modules (unit, geen e2e)
#   RACE=1 ./scripts/test.sh       # met -race
#   IMAGE=golang:1.24.6 ./scripts/test.sh   # andere basis-image
#
set -euo pipefail

IMAGE="${IMAGE:-golang:1.24.6}"
TEST_IMAGE="openftv-test:local"   # helper-image = IMAGE + curl/procps (gecached)

# ───────────────────────── host-zijde ─────────────────────────
# Buiten de container: helper-image bouwen (indien nodig) en re-exec in docker.
if [[ -z "${OPENFTV_IN_CONTAINER:-}" ]]; then
  REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

  if ! command -v docker >/dev/null 2>&1; then
    echo "FOUT: docker niet gevonden — dit script draait alle tests in een container." >&2
    exit 1
  fi

  if ! docker image inspect "$TEST_IMAGE" >/dev/null 2>&1; then
    echo ">> eenmalig test-image bouwen ($TEST_IMAGE op basis van $IMAGE)"
    docker build -t "$TEST_IMAGE" - <<EOF
FROM ${IMAGE}
RUN apt-get update \
 && apt-get install -y --no-install-recommends curl procps ca-certificates \
 && rm -rf /var/lib/apt/lists/*
EOF
  fi

  TTY=()
  [[ -t 1 ]] && TTY=(-t)

  echo ">> tests draaien in container (cache: volumes openftv-gomod / openftv-gocache)"
  exec docker run --rm -i ${TTY[@]+"${TTY[@]}"} \
    -e OPENFTV_IN_CONTAINER=1 \
    -e RACE="${RACE:-}" \
    -v "${REPO_ROOT}:/src" \
    -v openftv-gomod:/go/pkg/mod \
    -v openftv-gocache:/root/.cache/go-build \
    -w /src \
    "$TEST_IMAGE" \
    bash /src/scripts/test.sh "$@"
fi

# ──────────────────────── container-zijde ─────────────────────
cd /src

TESTFLAGS=(-cover -timeout=120s)
[[ -n "${RACE:-}" ]] && TESTFLAGS+=(-race)

# Modules uit go.work halen — single source of truth voor de workspace.
mapfile -t MODULES < <(awk '/^use \(/{f=1;next} /^\)/{f=0} f{gsub(/[ \t]/,"");print}' go.work | grep -v '^$')

run_unit=1
run_e2e=1
FILTER=()
for arg in "$@"; do
  case "$arg" in
    --unit) run_e2e=0 ;;
    --e2e)  run_unit=0 ;;
    *)      FILTER+=("$arg"); run_e2e=0 ;;   # expliciete modules → geen e2e
  esac
done

fails=()

if [[ $run_unit -eq 1 ]]; then
  targets=("${MODULES[@]}")
  [[ ${#FILTER[@]} -gt 0 ]] && targets=("${FILTER[@]}")
  for m in "${targets[@]}"; do
    if [[ ! -f "$m/go.mod" ]]; then
      echo "-- skip $m (geen go.mod)"
      continue
    fi
    echo "==> unit: go test ${TESTFLAGS[*]} ./...  ($m)"
    if ! (cd "$m" && go test "${TESTFLAGS[@]}" ./...); then
      fails+=("unit:$m")
    fi
  done
fi

if [[ $run_e2e -eq 1 ]]; then
  for e in e2e/*/test.sh; do
    [[ -f "$e" ]] || continue
    echo "==> e2e: $e"
    if ! timeout 300 bash "$e"; then
      fails+=("e2e:$e")
    fi
  done
fi

echo
if [[ ${#fails[@]} -eq 0 ]]; then
  echo "✅ alle tests geslaagd"
else
  echo "❌ ${#fails[@]} onderdeel(en) gefaald:"
  printf '   - %s\n' "${fails[@]}"
  exit 1
fi
