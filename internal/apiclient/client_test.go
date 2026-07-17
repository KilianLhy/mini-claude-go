package apiclient_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/KilianLhy/mini-claude-go/internal/api"
	"github.com/KilianLhy/mini-claude-go/internal/apiclient"
	"github.com/KilianLhy/mini-claude-go/internal/shared"
)

func TestClientAgainstServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := api.NewServer(api.NewMemoryStore(), []byte("test-secret"))
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	ctx := context.Background()
	c := apiclient.New(ts.URL, "")

	resp, err := c.Register(ctx, "hugo@example.com", "supersecret")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	c.SetToken(resp.Token)

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
