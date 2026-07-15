// Package shared defines the data contract exchanged between the mini-claude
// CLI and the mini-claude server. Both binaries import these types so the
// wire format is declared exactly once and can never drift out of sync.
//
// Nothing here depends on Bubble Tea, Gin, or the database: it is pure data.
package shared

import "time"

// Roles used in a chat message. They match the OpenAI/Ollama chat API.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Config holds the user-configurable settings. It is persisted locally in
// config.json and synced to the server. Field names use snake_case JSON tags
// so the same document is readable by the CLI, the server, and a human.
type Config struct {
	BaseURL      string  `json:"base_url"`
	Model        string  `json:"model"`
	Temperature  float64 `json:"temperature"`
	SystemPrompt string  `json:"system_prompt"`
	Theme        string  `json:"theme"`
}

// Message is a single chat turn. The JSON tags are kept identical to the
// Ollama request format so a Message can be sent to the model as-is.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// State is the persisted application state: the conversation history. It is
// stored locally in state.json and synced to the server.
type State struct {
	Messages []Message `json:"messages"`
}

// --- API envelopes -------------------------------------------------------

// RegisterRequest is the body of POST /auth/register.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the body of POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned by the auth endpoints on success. The token is a
// JWT the CLI stores and sends as "Authorization: Bearer <token>".
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// DataPayload is the current config+state for a user. It is the body of
// GET/PUT /me/data and the shape pushed by POST /me/export.
type DataPayload struct {
	Config    Config    `json:"config"`
	State     State     `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BackupSummary describes one historical backup, without its full contents.
// Returned as a list by GET /me/backups.
type BackupSummary struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorResponse is the uniform error body returned by every failing endpoint.
type ErrorResponse struct {
	Error string `json:"error"`
}

// --- Routes --------------------------------------------------------------
//
// Path constants shared by the server (route registration) and the CLI
// (request building), so a rename can never desynchronize the two sides.
const (
	RouteRegister   = "/auth/register"
	RouteLogin      = "/auth/login"
	RouteData       = "/me/data"    // GET, PUT
	RouteExport     = "/me/export"  // POST
	RouteImport     = "/me/import"  // POST
	RouteBackups    = "/me/backups" // GET (list)
	RouteBackupByID = "/me/backups/:id"
)
