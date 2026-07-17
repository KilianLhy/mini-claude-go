// Command server runs the mini-claude sync API.
//
// Configuration comes from the environment:
//
//	PORT          HTTP port to listen on (default 8080)
//	DATABASE_URL  PostgreSQL DSN. If empty, an in-memory store is used
//	              (data is lost on restart) — handy for local testing.
//	JWT_SECRET    Secret used to sign auth tokens (required in production).
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KilianLhy/mini-claude-go/internal/api"
)

func main() {
	// Run Gin in release mode unless GIN_MODE says otherwise (quiet logs, no
	// debug warnings in production).
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	port := getenv("PORT", "8080")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-insecure-secret-change-me"
		log.Println("WARNING: JWT_SECRET not set, using an insecure development secret")
	}

	store := openStore(ctx)
	defer store.Close()

	srv := api.NewServer(store, []byte(secret))
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Shut down gracefully on signal.
	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	log.Printf("mini-claude server listening on :%s", port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

// openStore returns a PostgreSQL store when DATABASE_URL is set, otherwise an
// in-memory fallback so the server can run with zero setup.
func openStore(ctx context.Context) api.Store {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("DATABASE_URL not set, using in-memory store (data will not persist)")
		return api.NewMemoryStore()
	}
	store, err := api.NewPostgresStore(ctx, dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	log.Println("connected to PostgreSQL")
	return store
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
