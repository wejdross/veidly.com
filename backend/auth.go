package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

// bcryptCost is the cost factor for password hashing
// Lower for tests (4), higher for production (14)
var bcryptCost = 14

func init() {
	// Use lower bcrypt cost for tests to speed them up
	if isTestMode() {
		bcryptCost = 4
	}
}

type Claims struct {
	UserID  int    `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(user User) (string, error) {
	claims := Claims{
		UserID:  user.ID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours (reduced from 7 days for security)
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// initJWTFromEnv loads the JWT secret from the environment.
// It fails if JWT_SECRET is not defined, empty, or too short.
func initJWTFromEnv() error {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	if secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}

	// Enforce minimum length for security (NIST recommends 256 bits = 43 base64 chars for HMAC-SHA256)
	if len(secret) < 43 {
		return fmt.Errorf("JWT_SECRET must be at least 43 characters long (got %d characters). Generate with: openssl rand -base64 43", len(secret))
	}

	jwtSecret = []byte(secret)
	return nil
}

// Middleware to require authentication
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if user is blocked and get email verification status
		var isBlocked, emailVerified bool
		err = db.QueryRow("SELECT is_blocked, email_verified FROM users WHERE id = ?", claims.UserID).Scan(&isBlocked, &emailVerified)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			c.Abort()
			return
		}

		if isBlocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "User account is blocked"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("is_admin", claims.IsAdmin)
		c.Set("email_verified", emailVerified)

		c.Next()
	}
}

// Optional auth middleware - extracts user info if token present, but doesn't require it
func optionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth token, continue without setting user context
			c.Next()
			return
		}

		// Format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without setting user context
			c.Next()
			return
		}

		token := parts[1]
		claims, err := validateToken(token)
		if err != nil {
			// Invalid token, continue without setting user context
			c.Next()
			return
		}

		// Get user info from database
		var isBlocked, emailVerified bool
		err = db.QueryRow("SELECT is_blocked, email_verified FROM users WHERE id = ?", claims.UserID).Scan(&isBlocked, &emailVerified)
		if err != nil || isBlocked {
			// User not found or blocked, continue without setting user context
			c.Next()
			return
		}

		// Set user info in context for privacy filtering
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("is_admin", claims.IsAdmin)
		c.Set("email_verified", emailVerified)

		c.Next()
	}
}

// Middleware to require admin privileges
func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
