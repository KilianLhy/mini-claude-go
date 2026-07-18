# mini-claude

A fast, private TUI chat client for self-hosted LLMs, written in Go — with a
companion sync server so your conversations and settings follow you across
machines.

Talk to a local language model from your terminal. Inference stays on your
machine (via [Ollama](https://ollama.com)); an optional account lets you back up
and restore your config and history from your own server.

## Features

- **Terminal UI** built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) — streaming responses, multi-turn chat, polished rendering.
- **Themes** — switch color palettes from a settings screen (`/settings`).
- **Local persistence** — settings and history saved as JSON in your config directory.
- **Account & sync** — sign in (`/login`), then push/pull your config and history to the server (`/export`, `/import`), with server-side backups.
- **Self-hostable server** — REST API (Gin) backed by PostgreSQL (JSONB).
- **Monitoring** — Prometheus metrics + a provisioned Grafana dashboard.

## Architecture

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

The CLI works fully offline (chat via local Ollama). The server is an optional
sync layer: sign in to back up and restore your data.

## Install

Via Homebrew (macOS / Linux):

```bash
brew tap KilianLhy/tap
brew install mini-claude
mini-claude
```

Or download a binary for Windows / macOS / Linux from the
[GitHub releases](https://github.com/KilianLhy/mini-claude-go/releases).

## Prerequisites for chatting

The CLI needs a running [Ollama](https://ollama.com) with a model pulled:

```bash
ollama pull llama3.2:3b
```

## Commands

| Command | Action |
|---|---|
| `/help` | show all commands |
| `/model` | pick or switch the model |
| `/settings` | change the color theme |
| `/login` · `/register` | sign in / create an account |
| `/logout` | sign out |
| `/export` | push config + history to the server |
| `/import` | pull config + history from the server |
| `/clear` | start a fresh conversation |
| `/quit` | exit |

`enter` sends, `ctrl+j` inserts a newline.

## Configuration

Settings live in `~/.config/mini-claude/config.json` (and `state.json` for
history). Everything can be overridden by environment variables:

| Variable | Default | Meaning |
|---|---|---|
| `MINI_CLAUDE_URL` | `http://localhost:11434` | Ollama base URL |
| `MINI_CLAUDE_MODEL` | `llama3.2:3b` | default model |
| `MINI_CLAUDE_TEMPERATURE` | `0.7` | sampling temperature |
| `MINI_CLAUDE_SYSTEM` | — | system prompt |
| `MINI_CLAUDE_THEME` | `claude` | color theme |
| `MINI_CLAUDE_SERVER` | hosted URL | sync server URL |

Precedence: environment variable > `config.json` > built-in default.

## Running the server locally

Bring up the app (Postgres + server) plus Ollama and the monitoring stack:

```bash
docker compose up -d --build
```

- Server: <http://localhost:8080> (`/health`, `/metrics`)
- Grafana: <http://localhost:3000> (admin / admin) — the **mini-claude** dashboard is auto-provisioned
- Prometheus: <http://localhost:9090>
- Docs site: <http://localhost:1313> (Hugo)

Point the CLI at the local server for testing:

```bash
MINI_CLAUDE_SERVER=http://localhost:8080 go run ./cmd/tui
```

Without Docker, the server runs with an in-memory store (no database needed):

```bash
go run ./cmd/server
```

## Server API

| Method & path | Description |
|---|---|
| `POST /auth/register` · `POST /auth/login` | create account / sign in → JWT |
| `GET` / `PUT /me/data` | read / update current config + state |
| `POST /me/export` | push data and create a backup |
| `POST /me/import` | pull current data |
| `GET /me/backups` · `GET /me/backups/:id` | list / fetch backups |
| `GET /health` · `GET /metrics` | health check / Prometheus metrics |

Passwords are hashed with bcrypt; `/me/*` routes require a Bearer JWT. Config
and state are stored as PostgreSQL JSONB.

## Development

```bash
make build              # build CLI + server
make test               # run tests
make run-server         # run the server
make run-tui            # run the TUI
make build-server-linux # cross-compile the server for Linux (deployment)
```

Deployment guide (alwaysdata): [docs/DEPLOY.md](./docs/DEPLOY.md).

## Stack

Go · Bubble Tea / Lip Gloss / Bubbles · Gin · PostgreSQL (pgx) · JWT · bcrypt ·
Ollama · Docker Compose · Prometheus / Grafana · GoReleaser · GitHub Actions.

## Authors

Hugo Stawiarski · Kilian Lahaye · Moustapha yaya Sow

## License

MIT — see [LICENSE](./LICENSE).
