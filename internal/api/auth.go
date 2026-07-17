package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const tokenTTL = 24 * time.Hour

const contextUserID = "userID"

func hashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(h), nil
}

func checkPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func issueToken(secret []byte, userID string) (token string, expiresAt time.Time, err error) {
	expiresAt = time.Now().Add(tokenTTL)
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

func parseToken(secret []byte, tokenStr string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}
	if claims.Subject == "" {
		return "", errors.New("token has no subject")
	}
	return claims.Subject, nil
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		prefix := "Bearer "
		if !strings.HasPrefix(header, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorBody("missing bearer token"))
			return
		}
		userID, err := parseToken(s.jwtSecret, strings.TrimPrefix(header, prefix))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorBody("invalid or expired token"))
			return
		}
		c.Set(contextUserID, userID)
		c.Next()
	}
}
