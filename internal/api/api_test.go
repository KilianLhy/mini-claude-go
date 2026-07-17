package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/KilianLhy/mini-claude-go/internal/shared"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewServer(NewMemoryStore(), []byte("test-secret")).Router()
}

// do sends a JSON request and returns the recorder. token may be empty.
func do(t *testing.T, r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func registerAndToken(t *testing.T, r *gin.Engine, email, pw string) string {
	t.Helper()
	rec := do(t, r, http.MethodPost, shared.RouteRegister, "", shared.RegisterRequest{Email: email, Password: pw})
	if rec.Code != http.StatusOK {
		t.Fatalf("register: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var resp shared.AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode auth response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected a token")
	}
	return resp.Token
}

func TestFullSyncFlow(t *testing.T) {
	r := newTestRouter()
	token := registerAndToken(t, r, "hugo@example.com", "supersecret")

	// No data yet: empty payload.
	rec := do(t, r, http.MethodGet, shared.RouteData, token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get data: want 200, got %d", rec.Code)
	}

	// Push config + state.
	payload := shared.DataPayload{
		Config: shared.Config{Model: "llama3.2:3b", Theme: "midnight"},
		State:  shared.State{Messages: []shared.Message{{Role: shared.RoleUser, Content: "hi"}}},
	}
	rec = do(t, r, http.MethodPut, shared.RouteData, token, payload)
	if rec.Code != http.StatusOK {
		t.Fatalf("put data: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	// Read it back.
	rec = do(t, r, http.MethodGet, shared.RouteData, token, nil)
	var got shared.DataPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if got.Config.Theme != "midnight" || len(got.State.Messages) != 1 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	// Export creates a backup.
	rec = do(t, r, http.MethodPost, shared.RouteExport, token, payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("export: want 201, got %d (%s)", rec.Code, rec.Body.String())
	}

	// Backups list has one entry.
	rec = do(t, r, http.MethodGet, shared.RouteBackups, token, nil)
	var backups []shared.BackupSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &backups); err != nil {
		t.Fatalf("decode backups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("want 1 backup, got %d", len(backups))
	}
}

func TestAuthRequired(t *testing.T) {
	r := newTestRouter()
	rec := do(t, r, http.MethodGet, shared.RouteData, "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 without token, got %d", rec.Code)
	}
	rec = do(t, r, http.MethodGet, shared.RouteData, "not-a-jwt", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 with bad token, got %d", rec.Code)
	}
}

func TestRegisterValidation(t *testing.T) {
	r := newTestRouter()

	// Password too short.
	rec := do(t, r, http.MethodPost, shared.RouteRegister, "", shared.RegisterRequest{Email: "a@b.com", Password: "short"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for short password, got %d", rec.Code)
	}

	// Duplicate email.
	registerAndToken(t, r, "dup@example.com", "supersecret")
	rec = do(t, r, http.MethodPost, shared.RouteRegister, "", shared.RegisterRequest{Email: "dup@example.com", Password: "supersecret"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("want 409 for duplicate email, got %d", rec.Code)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	r := newTestRouter()
	registerAndToken(t, r, "login@example.com", "supersecret")

	rec := do(t, r, http.MethodPost, shared.RouteLogin, "", shared.LoginRequest{Email: "login@example.com", Password: "wrongpass"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 for wrong password, got %d", rec.Code)
	}

	rec = do(t, r, http.MethodPost, shared.RouteLogin, "", shared.LoginRequest{Email: "login@example.com", Password: "supersecret"})
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 for correct login, got %d (%s)", rec.Code, rec.Body.String())
	}
}
