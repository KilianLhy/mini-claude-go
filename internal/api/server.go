package api

import (
	"github.com/gin-gonic/gin"

	"github.com/KilianLhy/mini-claude-go/internal/shared"
)

// Server holds the API dependencies and builds the router.
type Server struct {
	store     Store
	jwtSecret []byte
}

// NewServer wires a Server with its storage backend and JWT signing secret.
func NewServer(store Store, jwtSecret []byte) *Server {
	return &Server{store: store, jwtSecret: jwtSecret}
}

// Router builds the Gin engine with all routes registered.
func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), metricsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Prometheus scrape endpoint.
	r.GET("/metrics", metricsHandler())

	// Public auth endpoints.
	r.POST(shared.RouteRegister, s.handleRegister)
	r.POST(shared.RouteLogin, s.handleLogin)

	// Protected endpoints (require a valid JWT).
	auth := r.Group("/", s.authMiddleware())
	auth.GET(shared.RouteData, s.handleGetData)
	auth.PUT(shared.RouteData, s.handlePutData)
	auth.POST(shared.RouteExport, s.handleExport)
	auth.POST(shared.RouteImport, s.handleImport)
	auth.GET(shared.RouteBackups, s.handleListBackups)
	auth.GET(shared.RouteBackupByID, s.handleGetBackup)

	return r
}

// errorBody builds the uniform error response body.
func errorBody(msg string) shared.ErrorResponse {
	return shared.ErrorResponse{Error: msg}
}
