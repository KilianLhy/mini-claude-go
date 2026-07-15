BINDIR  := bin
LDFLAGS := -s -w

.PHONY: help build build-cli build-server build-server-linux test vet fmt \
        run-server run-tui docker-up docker-down clean

help:
	@echo "mini-claude — make targets:"
	@echo "  build              build CLI and server for the host"
	@echo "  build-cli          build the TUI client"
	@echo "  build-server       build the server for the host"
	@echo "  build-server-linux cross-compile the server to Linux amd64 (for alwaysdata)"
	@echo "  test               run all tests"
	@echo "  vet                run go vet"
	@echo "  fmt                format the code"
	@echo "  run-server         run the server locally"
	@echo "  run-tui            run the TUI locally"
	@echo "  docker-up          start Postgres + server with Docker Compose"
	@echo "  clean              remove build artifacts"

build: build-cli build-server

build-cli:
	go build -ldflags="$(LDFLAGS)" -o $(BINDIR)/mini-claude ./cmd/tui

build-server:
	go build -ldflags="$(LDFLAGS)" -o $(BINDIR)/mini-claude-server ./cmd/server

## Cross-compile the server as a static Linux amd64 binary for deployment.
build-server-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="$(LDFLAGS)" -o $(BINDIR)/mini-claude-server-linux-amd64 ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

run-server:
	go run ./cmd/server

run-tui:
	go run ./cmd/tui

docker-up:
	docker compose up --build

docker-down:
	docker compose down

clean:
	rm -rf $(BINDIR) dist
