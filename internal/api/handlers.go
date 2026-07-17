package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/KilianLhy/mini-claude-go/internal/shared"
)

// minPasswordLen is the minimum accepted password length.
const minPasswordLen = 8

// currentUser reads the user ID the auth middleware stored on the context.
func currentUser(c *gin.Context) string {
	v, _ := c.Get(contextUserID)
	id, _ := v.(string)
	return id
}

func (s *Server) handleRegister(c *gin.Context) {
	var req shared.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody("invalid JSON body"))
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		c.JSON(http.StatusBadRequest, errorBody("a valid email is required"))
		return
	}
	if len(req.Password) < minPasswordLen {
		c.JSON(http.StatusBadRequest, errorBody("password must be at least 8 characters"))
		return
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorBody("could not process password"))
		return
	}

	user, err := s.store.CreateUser(c.Request.Context(), req.Email, hash)
	if errors.Is(err, ErrEmailTaken) {
		authEvents.WithLabelValues("register", "failure").Inc()
		c.JSON(http.StatusConflict, errorBody("email already registered"))
		return
	}
	if err != nil {
		authEvents.WithLabelValues("register", "failure").Inc()
		c.JSON(http.StatusInternalServerError, errorBody("could not create account"))
		return
	}

	authEvents.WithLabelValues("register", "success").Inc()
	s.respondToken(c, user.ID)
}

func (s *Server) handleLogin(c *gin.Context) {
	var req shared.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorBody("invalid JSON body"))
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.store.UserByEmail(c.Request.Context(), req.Email)
	// Same response for unknown email and wrong password: don't leak which
	// accounts exist.
	if errors.Is(err, ErrNotFound) || (err == nil && !checkPassword(user.PasswordHash, req.Password)) {
		authEvents.WithLabelValues("login", "failure").Inc()
		c.JSON(http.StatusUnauthorized, errorBody("invalid email or password"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorBody("could not sign in"))
		return
	}

	authEvents.WithLabelValues("login", "success").Inc()
	s.respondToken(c, user.ID)
}

// respondToken issues a JWT for the user and writes the auth response.
func (s *Server) respondToken(c *gin.Context, userID string) {
	token, expiresAt, err := issueToken(s.jwtSecret, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorBody("could not issue token"))
		return
	}
	c.JSON(http.StatusOK, shared.AuthResponse{Token: token, ExpiresAt: expiresAt})
}

func (s *Server) handleGetData(c *gin.Context) {
	data, err := s.store.GetData(c.Request.Context(), currentUser(c))
	if errors.Is(err, ErrNotFound) {
		// Nothing synced yet: return an empty payload rather than a 404 so the
		// client can treat "no remote data" uniformly.
		c.JSON(http.StatusOK, shared.DataPayload{})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorBody("could not read data"))
		return
	}
	c.JSON(http.StatusOK, data)
}

func (s *Server) handlePutData(c *gin.Context) {
	var payload shared.DataPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorBody("invalid JSON body"))
		return
	}
	data, err := s.store.PutData(c.Request.Context(), currentUser(c), payload.Config, payload.State)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorBody("could not save data"))
		return
	}
	c.JSON(http.StatusOK, data)
}

// handleExport saves the posted config+state as the current data AND creates a
// timestamped backup.
func (s *Server) handleExport(c *gin.Context) {
	var payload shared.DataPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorBody("invalid JSON body"))
		return
	}
	userID := currentUser(c)
	if _, err := s.store.PutData(c.Request.Context(), userID, payload.Config, payload.State); err != nil {
		c.JSON(http.StatusInternalServerError, errorBody("could not save data"))
		return
	}
	backup, err := s.store.CreateBackup(c.Request.Context(), userID, payload.Config, payload.State)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorBody("could not create backup"))
		return
	}
	syncEvents.WithLabelValues("export").Inc()
	c.JSON(http.StatusCreated, backup)
}

// handleImport returns the user's current data (pull from server).
func (s *Server) handleImport(c *gin.Context) {
	syncEvents.WithLabelValues("import").Inc()
	s.handleGetData(c)
}

func (s *Server) handleListBackups(c *gin.Context) {
	backups, err := s.store.ListBackups(c.Request.Context(), currentUser(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorBody("could not list backups"))
		return
	}
	c.JSON(http.StatusOK, backups)
}

func (s *Server) handleGetBackup(c *gin.Context) {
	data, err := s.store.GetBackup(c.Request.Context(), currentUser(c), c.Param("id"))
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, errorBody("backup not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorBody("could not read backup"))
		return
	}
	c.JSON(http.StatusOK, data)
}
