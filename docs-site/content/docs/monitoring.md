---
title: Monitoring
weight: 7
---

# Monitoring

The server exposes application metrics on `/metrics` using the Prometheus Go
client. The local Docker stack ships a full observability setup:

- **Prometheus** scrapes the server, plus node-exporter (host) and cAdvisor
  (containers).
- **Grafana** auto-provisions a data source and a **mini-claude** dashboard.

## Metrics exposed

| Metric | Meaning |
|---|---|
| `miniclaude_http_requests_total` | requests by method, route, status |
| `miniclaude_http_request_duration_seconds` | request latency histogram |
| `miniclaude_auth_events_total` | registrations / logins (success & failure) |
| `miniclaude_sync_events_total` | exports / imports |

## Access

With the stack running (`docker compose up -d`):

- Grafana → <http://localhost:3000> (admin / admin) → dashboard
  **mini-claude — application & infra**
- Prometheus → <http://localhost:9090> (`/targets` to check scrape health)

Application panels fill once traffic hits the server; container CPU/memory
panels populate automatically via cAdvisor.
