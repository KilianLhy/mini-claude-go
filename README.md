# mini-claude

A fast, private TUI chat client for self-hosted LLMs, written in Go.

> Status: **WIP** — early scaffolding.

## Pitch

Talk to a local language model from your terminal. Nothing leaves your machine.
Streaming responses, multi-turn conversations, polished rendering. One tool, done well.

## Stack

- Go
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) / Lip Gloss / Bubbles / Glamour
- Standard library `net/http`
- [Ollama](https://ollama.com) as the inference backend (OpenAI-compatible API)
- Docker + Docker Compose
- GitLab CI/CD

## Quickstart (dev)

```bash
go run ./cmd/tui
```

More once the client actually does something.

## License

MIT — see [LICENSE](./LICENSE).
