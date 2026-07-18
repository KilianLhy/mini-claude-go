package api

import (
	"github.com/gin-gonic/gin"

	"github.com/KilianLhy/mini-claude-go/internal/shared"
)

type Server struct {
	store     Store
	jwtSecret []byte
}

func NewServer(store Store, jwtSecret []byte) *Server {
	return &Server{store: store, jwtSecret: jwtSecret}
}

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), metricsMiddleware(),
		securityHeadersMiddleware(), bodyLimitMiddleware(),
		rateLimitMiddleware(newRateLimiter(30, 60)))

	authLimiter := rateLimitMiddleware(newRateLimiter(0.2, 5))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/metrics", metricsHandler())

	r.POST(shared.RouteRegister, authLimiter, s.handleRegister)
	r.POST(shared.RouteLogin, authLimiter, s.handleLogin)

	auth := r.Group("/", s.authMiddleware())
	auth.GET(shared.RouteData, s.handleGetData)
	auth.PUT(shared.RouteData, s.handlePutData)
	auth.POST(shared.RouteExport, s.handleExport)
	auth.POST(shared.RouteImport, s.handleImport)
	auth.GET(shared.RouteBackups, s.handleListBackups)
	auth.GET(shared.RouteBackupByID, s.handleGetBackup)

	return r
}

func errorBody(msg string) shared.ErrorResponse {
	return shared.ErrorResponse{Error: msg}
}
