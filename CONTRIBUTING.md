# Contributing to mini-claude

Thanks for your interest. The project is in early scaffolding — issues, ideas, and PRs are welcome.

## Ground rules

- Keep the scope tight: one terminal chat client, done well.
- Local-first and private by default. No telemetry, no cloud round-trips.
- Match the existing Go style: `gofmt`, idiomatic packages, no dead code.

## Dev loop

```bash
go build ./...
go test ./...
go run ./cmd/tui
```

## Submitting changes

1. Fork or branch.
2. Run `gofmt` and `go test ./...`.
3. Open a merge request with a clear description.

Good first issues will be tagged once the MVP lands.
