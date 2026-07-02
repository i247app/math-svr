# Observability stack (Grafana · Loki · Prometheus · Tempo · Alloy)

This directory contains the local observability stack for `math-svr` and the
production hardening override. The Go app is instrumented for:

- **Metrics** — Prometheus (`client_golang`) on a dedicated listener.
- **Logs** — structured JSON tailed by Grafana Alloy → Loki.
- **Traces** — OpenTelemetry (OTLP/HTTP) → Tempo, spanning HTTP → DB → LLM.

Grafana is pre-provisioned with all three datasources and their cross-links
(log↔trace↔metric), plus a starter RED dashboard.

## Local usage

```bash
make obs-up      # start the stack
make obs-logs    # tail stack logs
make obs-down    # stop (keep volumes)
make obs-reset   # stop + wipe volumes
```

Then run the app on the host with (in `.env`):

```
OBS_TRACING_ENABLED=true
LOG_FILE=./logs/app.log
LOG_FILE_FORMAT=json
OBS_TRACE_SAMPLE_RATIO=1.0
```

Grafana → http://localhost:3000 (local: anonymous admin). Prometheus →
http://localhost:9090. See `.env.example` for every `OBS_*` / `LOG_*` key.

### Topology (local)

The app runs on the **host**; the stack runs in **Docker**:

- Prometheus scrapes the app at `host.docker.internal:9091` (the app's
  dedicated metrics listener, `OBS_METRICS_ADDR=:9091`).
- The app pushes traces to Tempo at `localhost:4318`.
- The app writes JSON logs to `../logs/app.log`, which Alloy tails (that dir is
  bind-mounted into the Alloy container).

## Production

Merge the hardening override:

```bash
GF_ADMIN_PASSWORD=... docker compose \
  -f docker/docker-compose.yml \
  -f docker/docker-compose.prod.yml up -d
```

Checklist beyond the override file:

1. **Secure `/metrics` (bind loopback).** Set `OBS_METRICS_ADDR=127.0.0.1:9091`
   so the endpoint is unreachable from the public internet. Scrape it with a
   node-local Prometheus/agent, or over a private network only. (There is no
   auth on `/metrics` by design — network isolation is the control.)
2. **Tracing endpoint.** Point `OBS_OTLP_ENDPOINT` at your real
   collector/agent (NOT `host.docker.internal`); set `OBS_OTLP_INSECURE=false`
   if it terminates TLS.
3. **Sampling.** Lower `OBS_TRACE_SAMPLE_RATIO` to `0.05`–`0.1`. `ParentBased`
   is applied, so upstream decisions are honoured. For "keep all error/slow
   traces", add tail-sampling in an OTel Collector (future work).
4. **Retention & volumes.** Tune retention in the config files (Prometheus flag
   in the override; Loki `retention_period`; Tempo `block_retention`) and back
   the named volumes with durable storage.
5. **Grafana.** The override disables anonymous access and requires
   `GF_ADMIN_PASSWORD`. Front it with TLS + SSO if exposed.
6. **Cardinality.** The `route` metric label / span name come from the route
   **classifier** (registered templates only; unknown paths → `unmatched`), so
   path-spraying cannot explode series. Keep it that way — never label metrics
   with raw paths, ids, emails, or other user-controlled values.

## What is (deliberately) NOT in telemetry

To avoid leaking secrets / PII:

- **DB spans** record the SQL statement with `?` placeholders only — bound
  values are never attached (the driver runs `InterpolateParams=true`).
- **LLM spans** record provider, model, and token counts — never the prompt
  or the response body.
- **HTTP spans** record the route template and `url.path` — never the query
  string. Keep this discipline when adding attributes.
