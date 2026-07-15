package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"gitlab.com/marseille-bb/mini-claude/internal/shared"
)

// PostgresStore persists users, current data, and backups in PostgreSQL.
// Config and State are stored as JSONB columns.
type PostgresStore struct {
	pool *pgxpool.Pool
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS user_data (
    user_id    BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    config     JSONB NOT NULL,
    state      JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS backups (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    config     JSONB NOT NULL,
    state      JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS backups_user_idx ON backups (user_id, created_at DESC);
`

// NewPostgresStore connects to the database, verifies the connection, and
// applies the schema (idempotent).
func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

// parseID converts a string user/backup ID to the bigint used by the database.
func parseID(id string) (int64, bool) {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func (p *PostgresStore) CreateUser(ctx context.Context, email, passwordHash string) (User, error) {
	var (
		id        int64
		createdAt time.Time
	)
	err := p.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at`,
		email, passwordHash,
	).Scan(&id, &createdAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return User{}, ErrEmailTaken
	}
	if err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return User{ID: strconv.FormatInt(id, 10), Email: email, PasswordHash: passwordHash, CreatedAt: createdAt}, nil
}

func (p *PostgresStore) UserByEmail(ctx context.Context, email string) (User, error) {
	var (
		id        int64
		hash      string
		createdAt time.Time
	)
	err := p.pool.QueryRow(ctx,
		`SELECT id, password_hash, created_at FROM users WHERE email = $1`, email,
	).Scan(&id, &hash, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("select user: %w", err)
	}
	return User{ID: strconv.FormatInt(id, 10), Email: email, PasswordHash: hash, CreatedAt: createdAt}, nil
}

func (p *PostgresStore) GetData(ctx context.Context, userID string) (shared.DataPayload, error) {
	uid, ok := parseID(userID)
	if !ok {
		return shared.DataPayload{}, ErrNotFound
	}
	var (
		cfgRaw, stRaw []byte
		updatedAt     time.Time
	)
	err := p.pool.QueryRow(ctx,
		`SELECT config, state, updated_at FROM user_data WHERE user_id = $1`, uid,
	).Scan(&cfgRaw, &stRaw, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.DataPayload{}, ErrNotFound
	}
	if err != nil {
		return shared.DataPayload{}, fmt.Errorf("select data: %w", err)
	}
	return decodePayload(cfgRaw, stRaw, updatedAt)
}

func (p *PostgresStore) PutData(ctx context.Context, userID string, cfg shared.Config, st shared.State) (shared.DataPayload, error) {
	uid, ok := parseID(userID)
	if !ok {
		return shared.DataPayload{}, ErrNotFound
	}
	cfgRaw, stRaw, err := encode(cfg, st)
	if err != nil {
		return shared.DataPayload{}, err
	}
	var updatedAt time.Time
	err = p.pool.QueryRow(ctx,
		`INSERT INTO user_data (user_id, config, state, updated_at)
		 VALUES ($1, $2, $3, now())
		 ON CONFLICT (user_id) DO UPDATE
		   SET config = EXCLUDED.config, state = EXCLUDED.state, updated_at = now()
		 RETURNING updated_at`,
		uid, cfgRaw, stRaw,
	).Scan(&updatedAt)
	if err != nil {
		return shared.DataPayload{}, fmt.Errorf("upsert data: %w", err)
	}
	return shared.DataPayload{Config: cfg, State: st, UpdatedAt: updatedAt}, nil
}

func (p *PostgresStore) CreateBackup(ctx context.Context, userID string, cfg shared.Config, st shared.State) (shared.BackupSummary, error) {
	uid, ok := parseID(userID)
	if !ok {
		return shared.BackupSummary{}, ErrNotFound
	}
	cfgRaw, stRaw, err := encode(cfg, st)
	if err != nil {
		return shared.BackupSummary{}, err
	}
	var (
		id        int64
		createdAt time.Time
	)
	err = p.pool.QueryRow(ctx,
		`INSERT INTO backups (user_id, config, state) VALUES ($1, $2, $3) RETURNING id, created_at`,
		uid, cfgRaw, stRaw,
	).Scan(&id, &createdAt)
	if err != nil {
		return shared.BackupSummary{}, fmt.Errorf("insert backup: %w", err)
	}
	return shared.BackupSummary{ID: strconv.FormatInt(id, 10), CreatedAt: createdAt}, nil
}

func (p *PostgresStore) ListBackups(ctx context.Context, userID string) ([]shared.BackupSummary, error) {
	uid, ok := parseID(userID)
	if !ok {
		return nil, ErrNotFound
	}
	rows, err := p.pool.Query(ctx,
		`SELECT id, created_at FROM backups WHERE user_id = $1 ORDER BY created_at DESC`, uid)
	if err != nil {
		return nil, fmt.Errorf("select backups: %w", err)
	}
	defer rows.Close()

	var out []shared.BackupSummary
	for rows.Next() {
		var (
			id        int64
			createdAt time.Time
		)
		if err := rows.Scan(&id, &createdAt); err != nil {
			return nil, fmt.Errorf("scan backup: %w", err)
		}
		out = append(out, shared.BackupSummary{ID: strconv.FormatInt(id, 10), CreatedAt: createdAt})
	}
	return out, rows.Err()
}

func (p *PostgresStore) GetBackup(ctx context.Context, userID, backupID string) (shared.DataPayload, error) {
	uid, ok := parseID(userID)
	bid, ok2 := parseID(backupID)
	if !ok || !ok2 {
		return shared.DataPayload{}, ErrNotFound
	}
	var (
		cfgRaw, stRaw []byte
		createdAt     time.Time
	)
	err := p.pool.QueryRow(ctx,
		`SELECT config, state, created_at FROM backups WHERE id = $1 AND user_id = $2`, bid, uid,
	).Scan(&cfgRaw, &stRaw, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.DataPayload{}, ErrNotFound
	}
	if err != nil {
		return shared.DataPayload{}, fmt.Errorf("select backup: %w", err)
	}
	return decodePayload(cfgRaw, stRaw, createdAt)
}

// Close releases the connection pool.
func (p *PostgresStore) Close() {
	p.pool.Close()
}

// encode marshals config and state to JSON for JSONB storage.
func encode(cfg shared.Config, st shared.State) (cfgRaw, stRaw []byte, err error) {
	if cfgRaw, err = json.Marshal(cfg); err != nil {
		return nil, nil, fmt.Errorf("encode config: %w", err)
	}
	if stRaw, err = json.Marshal(st); err != nil {
		return nil, nil, fmt.Errorf("encode state: %w", err)
	}
	return cfgRaw, stRaw, nil
}

// decodePayload unmarshals JSONB columns back into a DataPayload.
func decodePayload(cfgRaw, stRaw []byte, at time.Time) (shared.DataPayload, error) {
	var d shared.DataPayload
	if err := json.Unmarshal(cfgRaw, &d.Config); err != nil {
		return shared.DataPayload{}, fmt.Errorf("decode config: %w", err)
	}
	if err := json.Unmarshal(stRaw, &d.State); err != nil {
		return shared.DataPayload{}, fmt.Errorf("decode state: %w", err)
	}
	d.UpdatedAt = at
	return d, nil
}
