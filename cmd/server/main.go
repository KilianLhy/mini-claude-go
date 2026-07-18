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

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	port := getenv("PORT", "8080")
	dsn := os.Getenv("DATABASE_URL")
	secret := resolveSecret(dsn)

	store := openStore(ctx, dsn)
	defer store.Close()

	srv := api.NewServer(store, secret)
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

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

const minSecretLen = 32

func resolveSecret(dsn string) []byte {
	secret := os.Getenv("JWT_SECRET")
	if dsn != "" {
		if len(secret) < minSecretLen {
			log.Fatalf("JWT_SECRET must be set to at least %d characters in production (DATABASE_URL is set)", minSecretLen)
		}
		return []byte(secret)
	}
	if secret == "" {
		log.Println("WARNING: JWT_SECRET not set, using an insecure development secret (in-memory mode only)")
		return []byte("dev-insecure-secret-change-me")
	}
	return []byte(secret)
}

func openStore(ctx context.Context, dsn string) api.Store {
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
