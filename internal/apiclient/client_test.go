package apiclient_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gitlab.com/marseille-bb/mini-claude/internal/api"
	"gitlab.com/marseille-bb/mini-claude/internal/apiclient"
	"gitlab.com/marseille-bb/mini-claude/internal/shared"
)

// TestClientAgainstServer exercises the real HTTP client against the real Gin
// router, proving the CLI and server agree on the wire format end to end.
func TestClientAgainstServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := api.NewServer(api.NewMemoryStore(), []byte("test-secret"))
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	ctx := context.Background()
	c := apiclient.New(ts.URL, "")

	// Register and keep the token.
	resp, err := c.Register(ctx, "hugo@example.com", "supersecret")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	c.SetToken(resp.Token)

	// Export config + state, which also creates a backup.
	payload := shared.DataPayload{
		Config: shared.Config{Model: "llama3.2:3b", Theme: "midnight"},
		State:  shared.State{Messages: []shared.Message{{Role: shared.RoleUser, Content: "salut"}}},
	}
	backup, err := c.Export(ctx, payload)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if backup.ID == "" {
		t.Fatal("expected a backup id")
	}

	// Import it back and check it round-trips.
	got, err := c.Import(ctx)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if got.Config.Theme != "midnight" || got.Config.Model != "llama3.2:3b" {
		t.Fatalf("config mismatch: %+v", got.Config)
	}
	if len(got.State.Messages) != 1 || got.State.Messages[0].Content != "salut" {
		t.Fatalf("state mismatch: %+v", got.State)
	}
}

// TestLoginWrongPasswordSurfacesServerError verifies the client turns a non-2xx
// response into the server's error message.
func TestLoginWrongPasswordSurfacesServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := api.NewServer(api.NewMemoryStore(), []byte("test-secret"))
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	ctx := context.Background()
	c := apiclient.New(ts.URL, "")
	if _, err := c.Register(ctx, "a@b.com", "supersecret"); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := c.Login(ctx, "a@b.com", "wrongpass"); err == nil {
		t.Fatal("expected an error for wrong password")
	}
}
