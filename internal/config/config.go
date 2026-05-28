package config

import (
	"os"
	"strconv"
)

type Config struct {
	BaseURL      string
	Model        string
	Temperature  float64
	SystemPrompt string
}

func Load() Config {
	c := Config{
		BaseURL:      "http://localhost:11434",
		Model:        "llama3.2:3b",
		Temperature:  0.7,
		SystemPrompt: "",
	}
	if v := os.Getenv("MINI_CLAUDE_URL"); v != "" {
		c.BaseURL = v
	}
	if v := os.Getenv("MINI_CLAUDE_MODEL"); v != "" {
		c.Model = v
	}
	if v := os.Getenv("MINI_CLAUDE_TEMPERATURE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.Temperature = f
		}
	}
	if v := os.Getenv("MINI_CLAUDE_SYSTEM"); v != "" {
		c.SystemPrompt = v
	}
	return c
}
