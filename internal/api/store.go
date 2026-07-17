package api

import (
	"context"
	"errors"
	"time"

	"github.com/KilianLhy/mini-claude-go/internal/shared"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrEmailTaken = errors.New("email already registered")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash string) (User, error)
	UserByEmail(ctx context.Context, email string) (User, error)

	GetData(ctx context.Context, userID string) (shared.DataPayload, error)
	PutData(ctx context.Context, userID string, cfg shared.Config, st shared.State) (shared.DataPayload, error)

	CreateBackup(ctx context.Context, userID string, cfg shared.Config, st shared.State) (shared.BackupSummary, error)
	ListBackups(ctx context.Context, userID string) ([]shared.BackupSummary, error)
	GetBackup(ctx context.Context, userID, backupID string) (shared.DataPayload, error)

	Close()
}
