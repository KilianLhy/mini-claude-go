// Package store persists the CLI's configuration and state as JSON files in
// the user's config directory (e.g. ~/.config/mini-claude on Linux,
// %AppData%\mini-claude on Windows, ~/Library/Application Support on macOS).
//
// It is deliberately small: it knows how to locate the directory and how to
// read/write JSON atomically. It does not know what a Config or a State means
// beyond the shared contract types.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gitlab.com/marseille-bb/mini-claude/internal/shared"
)

const (
	appDir     = "mini-claude"
	configFile = "config.json"
	stateFile  = "state.json"
)

// Dir returns the mini-claude config directory, creating it if needed.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate config dir: %w", err)
	}
	dir := filepath.Join(base, appDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}

// loadInto reads name from the config dir and unmarshals it onto v. When the
// file does not exist it returns found=false and no error, leaving v as-is
// (so callers can pre-fill v with defaults). A corrupt file returns an error.
func loadInto(name string, v any) (found bool, err error) {
	dir, err := Dir()
	if err != nil {
		return false, err
	}
	path := filepath.Join(dir, name)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return false, fmt.Errorf("parse %s: %w", path, err)
	}
	return true, nil
}

// save writes v as pretty JSON to name atomically (temp file + rename) so a
// crash mid-write can never leave a truncated file.
func save(name string, v any) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	path := filepath.Join(dir, name)
	tmp, err := os.CreateTemp(dir, name+".tmp-*")
	if err != nil {
		return fmt.Errorf("temp file for %s: %w", name, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once renamed
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tmpName, err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return fmt.Errorf("chmod %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}

// LoadConfigInto overlays any saved config.json onto cfg (partial merge: only
// keys present in the file are overwritten). Returns whether a file was found.
func LoadConfigInto(cfg *shared.Config) (found bool, err error) {
	return loadInto(configFile, cfg)
}

// SaveConfig writes the config to config.json.
func SaveConfig(cfg shared.Config) error {
	return save(configFile, cfg)
}

// LoadState reads state.json. A missing file yields an empty state and no
// error; a corrupt file yields an empty state and the error, so the caller
// can surface it while still starting cleanly.
func LoadState() (shared.State, error) {
	var st shared.State
	if _, err := loadInto(stateFile, &st); err != nil {
		return shared.State{}, err
	}
	return st, nil
}

// SaveState writes the state to state.json.
func SaveState(st shared.State) error {
	return save(stateFile, st)
}
