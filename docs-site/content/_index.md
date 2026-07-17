---
title: mini-claude
type: docs
---

# mini-claude

A fast, private TUI chat client for self-hosted LLMs, written in Go — with a
companion sync server so your conversations and settings follow you across
machines.

Talk to a local language model from your terminal. Inference stays on your
machine (via [Ollama](https://ollama.com)); an optional account lets you back
up and restore your config and history from your own server.

## Highlights

- **Terminal UI** (Bubble Tea) — streaming responses, multi-turn chat, themes.
- **Local-first** — settings and history stored as JSON in your config directory.
- **Account & sync** — sign in, then push/pull your data to a self-hosted server.
- **REST server** — Gin + PostgreSQL (JSONB), JWT auth, per-user backups.
- **Observability** — Prometheus metrics + a provisioned Grafana dashboard.

[**Get started →**](/docs/installation/)

---

Built by Hugo Stawiarski · Kilian Lahaye · Moustapha Sow — MIT licensed.
