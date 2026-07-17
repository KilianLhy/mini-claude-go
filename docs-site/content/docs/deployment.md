---
title: Deployment
weight: 6
---

# Deployment

The server is a single Go binary that talks to PostgreSQL. It is configured
entirely through environment variables:

| Variable | Meaning |
|---|---|
| `PORT` | HTTP port (provided automatically on most hosts) |
| `DATABASE_URL` | PostgreSQL DSN; empty ⇒ in-memory store (dev only) |
| `JWT_SECRET` | secret used to sign auth tokens |

The schema is created automatically on startup (idempotent migration).

## Cross-compile for Linux

```bash
make build-server-linux
# -> bin/mini-claude-server-linux-amd64  (static ELF binary)
```

## Hosting (alwaysdata)

The production server runs on [alwaysdata](https://www.alwaysdata.com), which
hosts both the Go binary and PostgreSQL on a free plan. In short:

1. Create a PostgreSQL database and note its connection string.
2. Cross-compile and upload the Linux binary over SSH.
3. Create a **User program** site pointing at the binary, with `DATABASE_URL`
   and `JWT_SECRET` set as environment variables.

The live instance answers at `https://hugostarte.alwaysdata.net/health`.

## Local (Docker)

```bash
docker compose up -d --build
```

Brings up PostgreSQL, the server, Ollama, and the monitoring stack in one command.
