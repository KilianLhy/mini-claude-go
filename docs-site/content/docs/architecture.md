---
title: Architecture
weight: 4
---

# Architecture

The CLI works fully offline (chat via local Ollama). The server is an optional
sync layer: sign in to back up and restore your data.

```
 TUI client (Go)                     Sync server (Go)            Storage
┌────────────────┐   HTTP/JSON     ┌──────────────────┐        ┌────────────┐
│  Bubble Tea    │ ───────────────▶│  Gin REST API    │───────▶│ PostgreSQL │
│  chat, auth,   │   (JWT auth)    │  auth, data,     │  JSONB │ users/data │
│  settings      │◀─────────────── │  export, backups │        │ backups    │
└───────┬────────┘                 └────────┬─────────┘        └────────────┘
        │ OpenAI-compatible                 │ /metrics
        ▼ HTTP (local)                       ▼
   ┌──────────┐                      ┌──────────────┐
   │  Ollama  │                      │ Prometheus + │
   │  (local) │                      │  Grafana     │
   └──────────┘                      └──────────────┘
```

## Code layout

```
cmd/tui/          CLI entry point
cmd/server/       server entry point
internal/ui/      Bubble Tea model, views, themes, auth screen
internal/client/  Ollama streaming client
internal/chat/    conversation history
internal/config/  configuration loading
internal/store/   local JSON persistence (config, state, credentials)
internal/apiclient/ HTTP client for the sync server
internal/api/     Gin server: handlers, auth, stores, metrics
internal/shared/  the JSON contract shared by CLI and server
```

## Design notes

- **Single source of truth for the wire format**: `internal/shared` defines the
  JSON types imported by both the CLI and the server, so they can never drift.
- **Pluggable storage**: the server talks to a `Store` interface, implemented by
  a PostgreSQL backend (production) and an in-memory backend (tests / zero-setup).
- **Offline-first**: a missing or corrupt local file falls back to defaults; the
  CLI never crashes when the server or Ollama is unreachable.
