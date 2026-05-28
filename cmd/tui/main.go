package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gitlab.com/marseille-bb/mini-claude/internal/chat"
	"gitlab.com/marseille-bb/mini-claude/internal/client"
	"gitlab.com/marseille-bb/mini-claude/internal/config"
)

func main() {
	cfg := config.Load()
	cli := client.New(cfg.BaseURL, cfg.Model, cfg.Temperature)
	history := chat.New(cfg.SystemPrompt)

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("mini-claude — model: %s @ %s\n", cfg.Model, cfg.BaseURL)
	fmt.Println("type a message, blank line to skip, Ctrl+D or Ctrl+C to quit")

	reader := bufio.NewReader(os.Stdin)
	for {
		if ctx.Err() != nil {
			return
		}
		fmt.Print("\n> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println()
				return
			}
			fmt.Fprintf(os.Stderr, "read: %v\n", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		history.Add(chat.RoleUser, line)

		tokens, errs := cli.Stream(ctx, history.Messages())
		var reply strings.Builder
		for tok := range tokens {
			fmt.Print(tok)
			reply.WriteString(tok)
		}
		fmt.Println()
		if err := <-errs; err != nil {
			fmt.Fprintf(os.Stderr, "stream: %v\n", err)
			continue
		}
		history.Add(chat.RoleAssistant, reply.String())
	}
}
