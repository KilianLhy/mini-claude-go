package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"gitlab.com/marseille-bb/mini-claude/internal/client"
	"gitlab.com/marseille-bb/mini-claude/internal/config"
	"gitlab.com/marseille-bb/mini-claude/internal/ui"
)

func main() {
	cfg := config.Load()
	cli := client.New(cfg.BaseURL, cfg.Model, cfg.Temperature)

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	model := ui.New(cfg, cli, ctx)
	prog := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
