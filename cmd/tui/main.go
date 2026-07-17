package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/KilianLhy/mini-claude-go/internal/client"
	"github.com/KilianLhy/mini-claude-go/internal/config"
	"github.com/KilianLhy/mini-claude-go/internal/ui"
)

// version is injected at build time by GoReleaser via -ldflags; it stays
// "dev" for local `go run`/`go build`.
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println("mini-claude", version)
			return
		}
	}

	cfg, err := config.Load()
	if err != nil {
		// Non-fatal: a corrupt config file falls back to defaults.
		fmt.Fprintln(os.Stderr, "config:", err)
	}
	cli := client.New(cfg.BaseURL, cfg.Model, cfg.Temperature)

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	model := ui.New(cfg, cli, ctx)
	prog := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
