package config

import (
	"os"
	"strconv"

	"gitlab.com/marseille-bb/mini-claude/internal/shared"
	"gitlab.com/marseille-bb/mini-claude/internal/store"
)

// Config is the shared contract type. Aliasing keeps a single definition used
// by the CLI, the store, and the server.
type Config = shared.Config

// DefaultTheme is applied when nothing else specifies one.
const DefaultTheme = "claude"

// Defaults returns the built-in configuration.
func Defaults() Config {
	return Config{
		BaseURL:      "http://localhost:11434",
		Model:        "llama3.2:3b",
		Temperature:  0.7,
		SystemPrompt: "",
		Theme:        DefaultTheme,
	}
}

// Load builds the effective configuration with precedence
// defaults < config.json < environment variables. It never fails hard: on a
// corrupt config file it returns the best config it can plus a non-nil error
// the caller may surface, so the app still starts.
func Load() (Config, error) {
	cfg := Defaults()

	// Overlay the saved file (partial merge). A missing file is fine.
	_, err := store.LoadConfigInto(&cfg)

	// Environment overrides win over the file.
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

	return cfg, err
}

// Save persists the configuration to config.json.
func Save(cfg Config) error {
	return store.SaveConfig(cfg)
}
