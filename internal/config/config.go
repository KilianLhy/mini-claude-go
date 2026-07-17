package config

import (
	"os"
	"strconv"

	"github.com/KilianLhy/mini-claude-go/internal/shared"
	"github.com/KilianLhy/mini-claude-go/internal/store"
)

type Config = shared.Config

const DefaultTheme = "claude"

const DefaultServerURL = "https://hugostarte.alwaysdata.net"

func Defaults() Config {
	return Config{
		BaseURL:      "http://localhost:11434",
		Model:        "llama3.2:3b",
		Temperature:  0.7,
		SystemPrompt: "",
		Theme:        DefaultTheme,
		ServerURL:    DefaultServerURL,
	}
}

func Load() (Config, error) {
	cfg := Defaults()

	_, err := store.LoadConfigInto(&cfg)

	if v := os.Getenv("MINI_CLAUDE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("MINI_CLAUDE_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("MINI_CLAUDE_TEMPERATURE"); v != "" {
		if f, perr := strconv.ParseFloat(v, 64); perr == nil {
			cfg.Temperature = f
		}
	}
	if v := os.Getenv("MINI_CLAUDE_SYSTEM"); v != "" {
		cfg.SystemPrompt = v
	}
	if v := os.Getenv("MINI_CLAUDE_THEME"); v != "" {
		cfg.Theme = v
	}
	if v := os.Getenv("MINI_CLAUDE_SERVER"); v != "" {
		cfg.ServerURL = v
	}

	return cfg, err
}

func Save(cfg Config) error {
	return store.SaveConfig(cfg)
}
