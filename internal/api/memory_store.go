package api

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/KilianLhy/mini-claude-go/internal/shared"
)

type MemoryStore struct {
	mu      sync.Mutex
	seq     int
	users   map[string]User
	byEmail map[string]string
	data    map[string]shared.DataPayload
	backups map[string][]storedBackup
}

type storedBackup struct {
	summary shared.BackupSummary
	payload shared.DataPayload
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:   map[string]User{},
		byEmail: map[string]string{},
		data:    map[string]shared.DataPayload{},
		backups: map[string][]storedBackup{},
	}
}

func (m *MemoryStore) nextID() string {
	m.seq++
	return strconv.Itoa(m.seq)
}

func (m *MemoryStore) CreateUser(_ context.Context, email, passwordHash string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.byEmail[email]; exists {
		return User{}, ErrEmailTaken
	}
	u := User{ID: m.nextID(), Email: email, PasswordHash: passwordHash, CreatedAt: time.Now()}
	m.users[u.ID] = u
	m.byEmail[email] = u.ID
	return u, nil
}

func (m *MemoryStore) UserByEmail(_ context.Context, email string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byEmail[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return m.users[id], nil
}

func (m *MemoryStore) GetData(_ context.Context, userID string) (shared.DataPayload, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.data[userID]
	if !ok {
		return shared.DataPayload{}, ErrNotFound
	}
	return d, nil
}

func (m *MemoryStore) PutData(_ context.Context, userID string, cfg shared.Config, st shared.State) (shared.DataPayload, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d := shared.DataPayload{Config: cfg, State: st, UpdatedAt: time.Now()}
	m.data[userID] = d
	return d, nil
}

func (m *MemoryStore) CreateBackup(_ context.Context, userID string, cfg shared.Config, st shared.State) (shared.BackupSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	summary := shared.BackupSummary{ID: m.nextID(), CreatedAt: time.Now()}
	m.backups[userID] = append(m.backups[userID], storedBackup{
		summary: summary,
		payload: shared.DataPayload{Config: cfg, State: st, UpdatedAt: summary.CreatedAt},
	})
	return summary, nil
}

func (m *MemoryStore) ListBackups(_ context.Context, userID string) ([]shared.BackupSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]shared.BackupSummary, 0, len(m.backups[userID]))
	for _, b := range m.backups[userID] {
		out = append(out, b.summary)
	}
	return out, nil
}

func (m *MemoryStore) GetBackup(_ context.Context, userID, backupID string) (shared.DataPayload, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, b := range m.backups[userID] {
		if b.summary.ID == backupID {
			return b.payload, nil
		}
	}
	return shared.DataPayload{}, ErrNotFound
}

func (m *MemoryStore) Close() {}
