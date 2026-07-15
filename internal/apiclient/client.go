// Package apiclient is the CLI-side HTTP client for the mini-claude sync
// server. It speaks the shared contract and is used by the TUI to register,
// log in, and push/pull the user's config and state.
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.com/marseille-bb/mini-claude/internal/shared"
)

// Client talks to the sync server. It holds an optional bearer token.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// New builds a client for baseURL with an optional existing token.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// SetToken updates the bearer token (after login) or clears it (logout).
func (c *Client) SetToken(token string) { c.token = token }

// Token returns the current bearer token ("" when logged out).
func (c *Client) Token() string { return c.token }

// doJSON performs a JSON request, decoding into out (may be nil). Non-2xx
// responses are turned into an error carrying the server's message.
func (c *Client) doJSON(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("contact server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		var e shared.ErrorResponse
		if json.Unmarshal(data, &e) == nil && e.Error != "" {
			return fmt.Errorf("%s", e.Error)
		}
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// Register creates an account and returns an auth token.
func (c *Client) Register(ctx context.Context, email, password string) (shared.AuthResponse, error) {
	var out shared.AuthResponse
	err := c.doJSON(ctx, http.MethodPost, shared.RouteRegister,
		shared.RegisterRequest{Email: email, Password: password}, &out)
	return out, err
}

// Login authenticates and returns an auth token.
func (c *Client) Login(ctx context.Context, email, password string) (shared.AuthResponse, error) {
	var out shared.AuthResponse
	err := c.doJSON(ctx, http.MethodPost, shared.RouteLogin,
		shared.LoginRequest{Email: email, Password: password}, &out)
	return out, err
}

// Export pushes the current config+state and creates a server-side backup.
func (c *Client) Export(ctx context.Context, payload shared.DataPayload) (shared.BackupSummary, error) {
	var out shared.BackupSummary
	err := c.doJSON(ctx, http.MethodPost, shared.RouteExport, payload, &out)
	return out, err
}

// Import pulls the current config+state from the server.
func (c *Client) Import(ctx context.Context) (shared.DataPayload, error) {
	var out shared.DataPayload
	err := c.doJSON(ctx, http.MethodPost, shared.RouteImport, nil, &out)
	return out, err
}
