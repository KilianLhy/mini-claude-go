package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/KilianLhy/mini-claude-go/internal/shared"
)

const (
	appDir          = "mini-claude"
	configFile      = "config.json"
	stateFile       = "state.json"
	credentialsFile = "credentials.json"
)

type Credentials struct {
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

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
	defer os.Remove(tmpName)
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

func LoadConfigInto(cfg *shared.Config) (found bool, err error) {
	return loadInto(configFile, cfg)
}

func SaveConfig(cfg shared.Config) error {
	return save(configFile, cfg)
}

func LoadState() (shared.State, error) {
	var st shared.State
	if _, err := loadInto(stateFile, &st); err != nil {
		return shared.State{}, err
	}
	return st, nil
}

func SaveState(st shared.State) error {
	return save(stateFile, st)
}

func LoadCredentials() (creds Credentials, found bool, err error) {
	found, err = loadInto(credentialsFile, &creds)
	return creds, found, err
}

func SaveCredentials(creds Credentials) error {
	return save(credentialsFile, creds)
}

func ClearCredentials() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(dir, credentialsFile)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
