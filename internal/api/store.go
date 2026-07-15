// Package api implements the mini-claude sync server: a small REST API (Gin)
// that stores each user's CLI configuration and state, plus timestamped
// backups. It depends only on the shared contract for wire types and on a
// Store interface for persistence, so the storage backend (PostgreSQL, or the
// in-memory one used in tests) is pluggable.
package api

import (
	"context"
	"errors"
	"time"

	"gitlab.com/marseille-bb/mini-claude/internal/shared"
)

// Sentinel errors the storage layer returns; handlers map them to HTTP codes.
var (
	ErrNotFound   = errors.New("not found")
	ErrEmailTaken = errors.New("email already registered")
)

// User is a stored account. PasswordHash is a bcrypt hash; the plaintext
// password is never stored and never leaves the auth handlers.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// Store is the persistence contract. Both the PostgreSQL backend and the
// in-memory test backend implement it.
type Store interface {
	// Accounts.
	CreateUser(ctx context.Context, email, passwordHash string) (User, error)
	UserByEmail(ctx context.Context, email string) (User, error)

	// Current synced data for a user. GetData returns ErrNotFound when the
	// user has never synced.
	GetData(ctx context.Context, userID string) (shared.DataPayload, error)
	PutData(ctx context.Context, userID string, cfg shared.Config, st shared.State) (shared.DataPayload, error)

	// Backups (history). CreateBackup snapshots the given data.
	CreateBackup(ctx context.Context, userID string, cfg shared.Config, st shared.State) (shared.BackupSummary, error)
	ListBackups(ctx context.Context, userID string) ([]shared.BackupSummary, error)
	GetBackup(ctx context.Context, userID, backupID string) (shared.DataPayload, error)

	// Close releases any resources (no-op for the in-memory store).
	Close()
}
