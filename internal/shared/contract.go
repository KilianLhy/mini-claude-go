package shared

import "time"

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

type Config struct {
	BaseURL      string  `json:"base_url"`
	Model        string  `json:"model"`
	Temperature  float64 `json:"temperature"`
	SystemPrompt string  `json:"system_prompt"`
	Theme        string  `json:"theme"`
	ServerURL    string  `json:"server_url"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type State struct {
	Messages []Message `json:"messages"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type DataPayload struct {
	Config    Config    `json:"config"`
	State     State     `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BackupSummary struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

const (
	RouteRegister   = "/auth/register"
	RouteLogin      = "/auth/login"
	RouteData       = "/me/data"
	RouteExport     = "/me/export"
	RouteImport     = "/me/import"
	RouteBackups    = "/me/backups"
	RouteBackupByID = "/me/backups/:id"
)
