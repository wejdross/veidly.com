package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitJWTFromEnv(t *testing.T) {
	// Save original JWT secret
	origSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", origSecret)

	// Test with valid custom secret (must be at least 43 chars for HMAC-SHA256 security)
	validSecret := "my-custom-secret-key-at-least-43-characters-long-for-security"
	os.Setenv("JWT_SECRET", validSecret)
	err := initJWTFromEnv()
	assert.NoError(t, err)
	assert.NotNil(t, jwtSecret)
	assert.Equal(t, []byte(validSecret), jwtSecret)

	// Test with empty secret (should return error)
	os.Unsetenv("JWT_SECRET")
	err = initJWTFromEnv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")

	// Test with secret that's too short
	os.Setenv("JWT_SECRET", "short")
	err = initJWTFromEnv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 43 characters")
}

func setupJWT() {
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-at-least-43-characters-long-for-security-testing")
	initJWTFromEnv()
}

func TestGenerateToken(t *testing.T) {
	setupJWT()

	// Test token generation
	user := User{
		ID:      123,
		Email:   "test@example.com",
		IsAdmin: false,
	}
	token, err := generateToken(user)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be validated
	claims, err := validateToken(token)
	require.NoError(t, err)
	assert.Equal(t, 123, claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, false, claims.IsAdmin)
}

func TestValidateToken(t *testing.T) {
	setupJWT()

	// Test 1: Valid token
	user := User{ID: 456, Email: "admin@example.com", IsAdmin: true}
	token, _ := generateToken(user)
	claims, err := validateToken(token)
	require.NoError(t, err)
	assert.Equal(t, 456, claims.UserID)
	assert.Equal(t, "admin@example.com", claims.Email)
	assert.True(t, claims.IsAdmin)

	// Test 2: Invalid token
	_, err = validateToken("invalid-token-string")
	assert.Error(t, err)

	// Test 3: Empty token
	_, err = validateToken("")
	assert.Error(t, err)
}

func TestAuthMiddleware(t *testing.T) {
	setupJWT()
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified, is_blocked)
		VALUES (?, ?, ?, ?, ?)
	`, "test@example.com", hashedPassword, "Test User", 1, 0)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	// Create blocked user
	result, err = testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified, is_blocked)
		VALUES (?, ?, ?, ?, ?)
	`, "blocked@example.com", hashedPassword, "Blocked User", 1, 1)
	require.NoError(t, err)
	blockedUserID, _ := result.LastInsertId()

	router := gin.New()
	router.GET("/protected", authMiddleware(), func(c *gin.Context) {
		userID := c.GetInt("user_id")
		c.JSON(200, gin.H{"user_id": userID, "message": "success"})
	})

	// Test 1: Valid token
	user := User{ID: int(userID), Email: "test@example.com", IsAdmin: false}
	validToken, _ := generateToken(user)
	req1, _ := http.NewRequest("GET", "/protected", nil)
	req1.Header.Set("Authorization", "Bearer "+validToken)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "success")

	// Test 2: Missing Authorization header
	req2, _ := http.NewRequest("GET", "/protected", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// Test 3: Invalid token format (missing Bearer)
	req3, _ := http.NewRequest("GET", "/protected", nil)
	req3.Header.Set("Authorization", "invalid-format")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusUnauthorized, w3.Code)

	// Test 4: Invalid token
	req4, _ := http.NewRequest("GET", "/protected", nil)
	req4.Header.Set("Authorization", "Bearer invalid-token")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusUnauthorized, w4.Code)

	// Test 5: Blocked user
	blockedUser := User{ID: int(blockedUserID), Email: "blocked@example.com", IsAdmin: false}
	blockedToken, _ := generateToken(blockedUser)
	req5, _ := http.NewRequest("GET", "/protected", nil)
	req5.Header.Set("Authorization", "Bearer "+blockedToken)
	w5 := httptest.NewRecorder()
	router.ServeHTTP(w5, req5)

	assert.Equal(t, http.StatusForbidden, w5.Code)
	assert.Contains(t, w5.Body.String(), "blocked")
}

func TestOptionalAuthMiddleware(t *testing.T) {
	setupJWT()
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified, is_blocked)
		VALUES (?, ?, ?, ?, ?)
	`, "test@example.com", hashedPassword, "Test User", 1, 0)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	router := gin.New()
	router.GET("/optional", optionalAuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			c.JSON(200, gin.H{"authenticated": true, "user_id": userID})
		} else {
			c.JSON(200, gin.H{"authenticated": false})
		}
	})

	// Test 1: With valid token
	user := User{ID: int(userID), Email: "test@example.com", IsAdmin: false}
	validToken, _ := generateToken(user)
	req1, _ := http.NewRequest("GET", "/optional", nil)
	req1.Header.Set("Authorization", "Bearer "+validToken)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), `"authenticated":true`)

	// Test 2: Without token (should still work)
	req2, _ := http.NewRequest("GET", "/optional", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), `"authenticated":false`)

	// Test 3: With invalid token (should still work, just not authenticated)
	req3, _ := http.NewRequest("GET", "/optional", nil)
	req3.Header.Set("Authorization", "Bearer invalid-token")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Contains(t, w3.Body.String(), `"authenticated":false`)
}

func TestAdminMiddleware(t *testing.T) {
	setupJWT()
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create regular user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified, is_admin)
		VALUES (?, ?, ?, ?, ?)
	`, "user@example.com", hashedPassword, "Regular User", 1, 0)
	require.NoError(t, err)
	regularUserID, _ := result.LastInsertId()

	// Create admin user
	result, err = testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified, is_admin)
		VALUES (?, ?, ?, ?, ?)
	`, "admin@example.com", hashedPassword, "Admin User", 1, 1)
	require.NoError(t, err)
	adminUserID, _ := result.LastInsertId()

	router := gin.New()
	router.GET("/admin", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin access granted"})
	})

	// Test 1: Admin user has access
	adminUser := User{ID: int(adminUserID), Email: "admin@example.com", IsAdmin: true}
	adminToken, _ := generateToken(adminUser)
	req1, _ := http.NewRequest("GET", "/admin", nil)
	req1.Header.Set("Authorization", "Bearer "+adminToken)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "admin access granted")

	// Test 2: Regular user is forbidden
	regularUser := User{ID: int(regularUserID), Email: "user@example.com", IsAdmin: false}
	userToken, _ := generateToken(regularUser)
	req2, _ := http.NewRequest("GET", "/admin", nil)
	req2.Header.Set("Authorization", "Bearer "+userToken)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusForbidden, w2.Code)
	assert.Contains(t, w2.Body.String(), "Admin privileges required")
}

func TestAuthMiddlewareNonExistentUser(t *testing.T) {
	setupJWT()
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	router := gin.New()
	router.GET("/protected", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Generate token for non-existent user ID
	nonExistentUser := User{ID: 99999, Email: "nonexistent@example.com", IsAdmin: false}
	nonExistentToken, _ := generateToken(nonExistentUser)
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+nonExistentToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}
