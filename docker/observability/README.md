# OpenFTV ADL observability stack

Serves the Inzicht **statistieken**- and **inzage**-endpoints from a real OpenTelemetry
backend (Grafana Loki) instead of the local write-ahead log (WAL). The ADL record is the
single source of truth end-to-end: the exact JSON that the WAL writes per line is also the
Loki log line, so nothing is lost on the way to the backend and back.

```
 PDP / Inzicht ──OTLP──▶ Collector ──OTLP/HTTP──▶ Loki ◀──query_range── Inzicht (LokiSource)
                                    └─(optional)─▶ OpenSearch            Grafana (dashboard)
```

## Components

| Service    | Image                                              | Port | Role                                            |
|------------|----------------------------------------------------|------|-------------------------------------------------|
| collector  | `otel/opentelemetry-collector-contrib:0.116.0`     | 4317/4318 | OTLP gRPC+HTTP in, fan-out to Loki (+ OpenSearch) |
| loki       | `grafana/loki:3.3.0`                               | 3100 | Stores ADL records as OTLP logs                 |
| grafana    | `grafana/grafana:11.4.0`                           | 3000 | Datasource + "ADL-beslissingen" dashboard       |
| opensearch | `opensearchproject/opensearch:2.18.0` (profile)   | 9200 | Optional second backend                         |

## Exporter choice: `otlphttp` → Loki native OTLP (not the `loki` exporter)

The standalone **`loki` exporter was deprecated and removed** from collector-contrib. The
current, forward-compatible path is the standard **`otlphttp`** exporter pointing at Loki's
**native OTLP** endpoint (`http://loki:3100/otlp`). Loki 3.x ingests OTLP logs directly and,
via `limits_config.otlp_config` (see `loki-config.yaml`), promotes selected attributes to
index labels. `collector-config.yaml` documents this.

## Label choice and cardinality (important)

Loki indexes **labels**; every distinct label-value combination is a separate stream, so
labels must be **low-cardinality**. The ADL record, however, carries high-cardinality and
personal fields. The split we make:

| Field                         | Where                         | Why                                              |
|-------------------------------|-------------------------------|--------------------------------------------------|
| `job` = `adl`                 | **label** (constant)          | non-empty stream selector                        |
| `event_name`                  | **label**                     | 5 fixed values — bounded                         |
| `decision` (permit/deny/unset)| **label**                     | 3 values — bounded                               |
| `service_name`                | **label**                     | handful of producing services — bounded          |
| `trace_id`, `span_id`         | log line (+ OTLP trace/span id) | **unbounded** — would explode streams          |
| `subject` / afnemer id, `doel`, obligations, request/response | **log line (record JSON)** | high-cardinality and/or personal — must not be a label |

`afnemer` and `doel` are therefore **not** labels. The Grafana dashboard extracts them at
query time with the LogQL `json` parser, and `query.LokiSource` filters them **client-side**
(exactly like `WALSource`) after decoding each line back into a `Record`. Consequence: even
if a Loki/OTLP version promotes labels differently, correctness is unaffected — only the
push-down efficiency changes, never the result set.

## Backend selection in Inzicht

`apps/inzicht` chooses the read-side `query.Source` via `INZICHT_BACKEND`
(**not** `INZICHT_SOURCE`, which is the producer identifier):

```
INZICHT_BACKEND=wal          # default — reads the local WAL (ADL_PATH)
INZICHT_BACKEND=loki         # + INZICHT_LOKI_URL=http://localhost:3100
                             #   [INZICHT_LOKI_ORG_ID, INZICHT_LOKI_USER, INZICHT_LOKI_PASSWORD]
INZICHT_BACKEND=opensearch   # + INZICHT_OPENSEARCH_URL=http://localhost:9200
                             #   [INZICHT_OPENSEARCH_INDEX (default "adl"), _USER, _PASSWORD]
```

Producers (PDP, Inzicht itself) flush to the collector by using the ADL OTLP **log** sink
(`adl.NewOTLPLogSink`, endpoint `http://localhost:4318`), whose Body is the verbatim record
JSON. The older `adl.NewOTLPSink` uses the OpenTelemetry **trace** pipeline; it now also
mirrors the full record as the `adl.record` span attribute so that path is lossless too, but
Loki ingestion uses the log sink.

## Run it

```sh
cd docker/observability
docker compose -f compose.yaml up -d                      # collector + loki + grafana
docker compose -f compose.yaml --profile opensearch up -d # also start OpenSearch
```

Grafana: http://localhost:3000 → dashboard **ADL / ADL-beslissingen** (anonymous admin).

## End-to-end walkthrough (curl)

`docker` cannot run in the build environment; this is the host-run verification. The whole
point is: **a PDP decision must be queryable through `GET /v1/statistieken` served from Loki.**

### 1. Emit an ADL decision as an OTLP log to the collector

Post an `ExportLogsServiceRequest` (OTLP/JSON) whose log Body is the record JSON. This is
exactly the shape `adl.NewOTLPLogSink` produces (see `EncodeOTLPLog`):

```sh
NOW_NS=$(( $(date +%s) * 1000000000 ))
curl -s http://localhost:4318/v1/logs -H 'Content-Type: application/json' -d @- <<JSON
{ "resourceLogs": [ { "resource": { "attributes": [
      { "key": "service.name", "value": { "stringValue": "pdp" } } ] },
  "scopeLogs": [ { "scope": { "name": "adl" }, "logRecords": [ {
    "timeUnixNano": "${NOW_NS}",
    "traceId": "0af7651916cd43dd8448eb211c80319c",
    "spanId": "b7ad6b7169203331",
    "attributes": [
      { "key": "event_name",   "value": { "stringValue": "adl.access_evaluation" } },
      { "key": "decision",     "value": { "stringValue": "permit" } },
      { "key": "service_name", "value": { "stringValue": "pdp" } } ],
    "body": { "stringValue": "{\"trace_id\":\"0af7651916cd43dd8448eb211c80319c\",\"span_id\":\"b7ad6b7169203331\",\"event_name\":\"adl.access_evaluation\",\"timestamp\":$(( $(date +%s) * 1000 )),\"status\":\"Ok\",\"resource\":{\"service.name\":\"pdp\"},\"body\":{\"adl.core.request\":{\"subject\":{\"id\":\"gemeente-amsterdam\"},\"action\":{\"name\":\"read\"},\"context\":{\"purpose\":\"huisvestingswet\"}},\"adl.core.response\":{\"decision\":true}}}" }
  } ] } ] } ] }
JSON
```

### 2. Confirm it landed in Loki

```sh
curl -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={job="adl", event_name="adl.access_evaluation", decision="permit"}' \
  --data-urlencode "start=$(( ($(date +%s) - 3600) * 1000000000 ))" \
  --data-urlencode "end=$(( ($(date +%s) + 60) * 1000000000 ))" | jq '.data.result[].values'
```

You should see the verbatim record JSON as the log line.

### 3. Serve `GET /v1/statistieken` from Loki

Run Inzicht against the Loki backend and query the aggregate endpoint:

```sh
INZICHT_BACKEND=loki INZICHT_LOKI_URL=http://localhost:3100 \
INZICHT_AUTH_DISABLED=true \
  go run ./apps/inzicht/cmd &

curl -s "http://localhost:8443/v1/statistieken?bucket=day" -H 'X-Verstrekker: rvig' | jq
```

The returned `statistieken[]` (permit/deny per afnemer/doel/periode, k-anonymised) are now
computed from records read out of Loki — not from the WAL. The same `LokiSource` backs the
inzage endpoints (`/v1/inzicht/verzoeken/{id}/resultaat`).

> Send ≥ `INZICHT_K` (default 5) decisions per (afnemer, doel) bucket, otherwise
> k-anonymity suppresses the bucket and `statistieken[]` is empty by design.

## OpenSearch branch

Enable the `logs/opensearch` pipeline in `collector-config.yaml` and the `opensearch` compose
profile. The record is written verbatim as the document `_source`, so
`query.OpenSearchSource` (`INZICHT_BACKEND=opensearch`) reconstructs it losslessly. Grafana
needs the `grafana-opensearch-datasource` plugin (see the commented datasource).
