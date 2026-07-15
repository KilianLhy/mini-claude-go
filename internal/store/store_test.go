package store

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/marseille-bb/mini-claude/internal/shared"
)

// redirect points os.UserConfigDir at a temp directory for the test.
func redirect(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	// Linux uses XDG_CONFIG_HOME; macOS/Windows use HOME/AppData. Set the
	// common ones so the test is portable.
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	t.Setenv("AppData", dir)
}

func TestConfigRoundTrip(t *testing.T) {
	redirect(t)

	want := shared.Config{
		BaseURL:     "http://example:1234",
		Model:       "qwen2.5",
		Temperature: 0.42,
		Theme:       "midnight",
	}
	if err := SaveConfig(want); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	got := shared.Config{}
	found, err := LoadConfigInto(&got)
	if err != nil {
		t.Fatalf("LoadConfigInto: %v", err)
	}
	if !found {
		t.Fatal("expected config file to be found")
	}
	if got != want {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, want)
	}
}

func TestConfigPartialMergeKeepsDefaults(t *testing.T) {
	redirect(t)

	// Simulate an older/partial config file with only one field set.
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, configFile), []byte(`{"model":"custom"}`), 0o600); err != nil {
		t.Fatalf("write partial: %v", err)
	}

	cfg := shared.Config{BaseURL: "default-url", Model: "default", Theme: "default"}
	if _, err := LoadConfigInto(&cfg); err != nil {
		t.Fatalf("LoadConfigInto: %v", err)
	}
	if cfg.Model != "custom" {
		t.Errorf("model not overlaid: got %q", cfg.Model)
	}
	if cfg.BaseURL != "default-url" || cfg.Theme != "default" {
		t.Errorf("untouched fields lost their defaults: %+v", cfg)
	}
}

func TestMissingConfigIsNotAnError(t *testing.T) {
	redirect(t)
	cfg := shared.Config{Model: "default"}
	found, err := LoadConfigInto(&cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected found=false when no file exists")
	}
	if cfg.Model != "default" {
		t.Error("defaults should be untouched when file is missing")
	}
}

func TestStateRoundTrip(t *testing.T) {
	redirect(t)

	want := shared.State{Messages: []shared.Message{
		{Role: shared.RoleUser, Content: "hello"},
		{Role: shared.RoleAssistant, Content: "hi there"},
	}}
	if err := SaveState(want); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	got, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if len(got.Messages) != len(want.Messages) {
		t.Fatalf("len mismatch: got %d want %d", len(got.Messages), len(want.Messages))
	}
	for i := range want.Messages {
		if got.Messages[i] != want.Messages[i] {
			t.Errorf("message %d mismatch: got %+v want %+v", i, got.Messages[i], want.Messages[i])
		}
	}
}

func TestCorruptStateReturnsError(t *testing.T) {
	redirect(t)
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, stateFile), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write corrupt: %v", err)
	}
	if _, err := LoadState(); err == nil {
		t.Fatal("expected an error for corrupt state file")
	}
}
