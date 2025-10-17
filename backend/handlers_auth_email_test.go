package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyEmail(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "test@example.com", hashedPassword, "Test User", 0)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	// Create verification token
	token := "test-verification-token-123"
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = testDB.Exec(`
		INSERT INTO email_verification_tokens (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`, userID, token, expiresAt)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/api/verify-email", VerifyEmail)

	// Test 1: Successful verification
	req, _ := http.NewRequest("GET", "/api/verify-email?token="+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Email verified successfully")

	// Verify user is now verified
	var emailVerified bool
	err = testDB.QueryRow("SELECT email_verified FROM users WHERE id = ?", userID).Scan(&emailVerified)
	require.NoError(t, err)
	assert.True(t, emailVerified)

	// Test 2: Token already used (should fail)
	req2, _ := http.NewRequest("GET", "/api/verify-email?token="+token, nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
	assert.Contains(t, w2.Body.String(), "Invalid or expired")

	// Test 3: Missing token
	req3, _ := http.NewRequest("GET", "/api/verify-email", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusBadRequest, w3.Code)
	assert.Contains(t, w3.Body.String(), "token is required")

	// Test 4: Invalid token
	req4, _ := http.NewRequest("GET", "/api/verify-email?token=invalid-token", nil)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusBadRequest, w4.Code)
}

func TestVerifyEmailExpiredToken(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "test@example.com", hashedPassword, "Test User", 0)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	// Create EXPIRED verification token
	token := "expired-token-123"
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	_, err = testDB.Exec(`
		INSERT INTO email_verification_tokens (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`, userID, token, expiresAt)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/api/verify-email", VerifyEmail)

	req, _ := http.NewRequest("GET", "/api/verify-email?token="+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "expired")
}

func TestResendVerificationEmail(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create unverified test user
	hashedPassword, _ := hashPassword("password123")
	_, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "unverified@example.com", hashedPassword, "Unverified User", 0)
	require.NoError(t, err)

	// Create already verified user
	_, err = testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "verified@example.com", hashedPassword, "Verified User", 1)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/api/resend-verification", ResendVerificationEmail)

	// Test 1: Resend for non-existent email (security: don't reveal if exists)
	reqBody1 := map[string]string{"email": "nonexistent@example.com"}
	jsonData1, _ := json.Marshal(reqBody1)
	req1, _ := http.NewRequest("POST", "/api/resend-verification", bytes.NewBuffer(jsonData1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "verification link has been sent")

	// Test 2: Resend for already verified user
	reqBody2 := map[string]string{"email": "verified@example.com"}
	jsonData2, _ := json.Marshal(reqBody2)
	req2, _ := http.NewRequest("POST", "/api/resend-verification", bytes.NewBuffer(jsonData2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
	assert.Contains(t, w2.Body.String(), "already verified")

	// Test 3: Invalid request (missing email)
	reqBody3 := map[string]string{}
	jsonData3, _ := json.Marshal(reqBody3)
	req3, _ := http.NewRequest("POST", "/api/resend-verification", bytes.NewBuffer(jsonData3))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusBadRequest, w3.Code)

	// Test 4: Invalid email format
	reqBody4 := map[string]string{"email": "invalid-email"}
	jsonData4, _ := json.Marshal(reqBody4)
	req4, _ := http.NewRequest("POST", "/api/resend-verification", bytes.NewBuffer(jsonData4))
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusBadRequest, w4.Code)

	// Test 5: Successful resend with email service not configured
	emailService = nil
	reqBody5 := map[string]string{"email": "unverified@example.com"}
	jsonData5, _ := json.Marshal(reqBody5)
	req5, _ := http.NewRequest("POST", "/api/resend-verification", bytes.NewBuffer(jsonData5))
	req5.Header.Set("Content-Type", "application/json")
	w5 := httptest.NewRecorder()
	router.ServeHTTP(w5, req5)

	assert.Equal(t, http.StatusServiceUnavailable, w5.Code)
	assert.Contains(t, w5.Body.String(), "Email service is not configured")
}

func TestForgotPassword(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("password123")
	_, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "test@example.com", hashedPassword, "Test User", 1)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/api/forgot-password", ForgotPassword)

	// Test 1: Without email service configured (should return 503)
	emailService = nil
	reqBody1 := ForgotPasswordRequest{Email: "test@example.com"}
	jsonData1, _ := json.Marshal(reqBody1)
	req1, _ := http.NewRequest("POST", "/api/forgot-password", bytes.NewBuffer(jsonData1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusServiceUnavailable, w1.Code)
	assert.Contains(t, w1.Body.String(), "Email service is not configured")

	// Test 2: Request for non-existent email (security: don't reveal if exists)
	reqBody2 := ForgotPasswordRequest{Email: "nonexistent@example.com"}
	jsonData2, _ := json.Marshal(reqBody2)
	req2, _ := http.NewRequest("POST", "/api/forgot-password", bytes.NewBuffer(jsonData2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "If the email exists")

	// Test 3: Invalid request (empty body)
	req3, _ := http.NewRequest("POST", "/api/forgot-password", bytes.NewBuffer([]byte("{}")))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusBadRequest, w3.Code)
}

func TestResetPassword(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("oldpassword123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "test@example.com", hashedPassword, "Test User", 1)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	// Create valid reset token
	validToken := "valid-reset-token-123"
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = testDB.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at, used)
		VALUES (?, ?, ?, ?)
	`, userID, validToken, expiresAt, 0)
	require.NoError(t, err)

	// Create expired reset token
	expiredToken := "expired-reset-token-123"
	expiredAt := time.Now().Add(-1 * time.Hour)
	_, err = testDB.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at, used)
		VALUES (?, ?, ?, ?)
	`, userID, expiredToken, expiredAt, 0)
	require.NoError(t, err)

	// Create already used token
	usedToken := "used-reset-token-123"
	usedExpiresAt := time.Now().Add(1 * time.Hour)
	_, err = testDB.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at, used)
		VALUES (?, ?, ?, ?)
	`, userID, usedToken, usedExpiresAt, 1)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/api/reset-password", ResetPassword)

	// Test 1: Successful password reset
	reqBody1 := ResetPasswordRequest{
		Token:       validToken,
		NewPassword: "newpassword123",
	}
	jsonData1, _ := json.Marshal(reqBody1)
	req1, _ := http.NewRequest("POST", "/api/reset-password", bytes.NewBuffer(jsonData1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "Password reset successfully")

	// Verify password was actually changed
	var newHashedPassword string
	err = testDB.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&newHashedPassword)
	require.NoError(t, err)
	assert.NotEqual(t, hashedPassword, newHashedPassword)

	// Test 2: Try to use the same token again (should fail)
	req2, _ := http.NewRequest("POST", "/api/reset-password", bytes.NewBuffer(jsonData1))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
	assert.Contains(t, w2.Body.String(), "already been used")

	// Test 3: Expired token
	reqBody3 := ResetPasswordRequest{
		Token:       expiredToken,
		NewPassword: "newpassword456",
	}
	jsonData3, _ := json.Marshal(reqBody3)
	req3, _ := http.NewRequest("POST", "/api/reset-password", bytes.NewBuffer(jsonData3))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusBadRequest, w3.Code)
	assert.Contains(t, w3.Body.String(), "expired")

	// Test 4: Invalid token
	reqBody4 := ResetPasswordRequest{
		Token:       "invalid-token",
		NewPassword: "newpassword789",
	}
	jsonData4, _ := json.Marshal(reqBody4)
	req4, _ := http.NewRequest("POST", "/api/reset-password", bytes.NewBuffer(jsonData4))
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusBadRequest, w4.Code)

	// Test 5: Invalid request (missing fields)
	reqBody5 := ResetPasswordRequest{Token: "some-token"}
	jsonData5, _ := json.Marshal(reqBody5)
	req5, _ := http.NewRequest("POST", "/api/reset-password", bytes.NewBuffer(jsonData5))
	req5.Header.Set("Content-Type", "application/json")
	w5 := httptest.NewRecorder()
	router.ServeHTTP(w5, req5)

	assert.Equal(t, http.StatusBadRequest, w5.Code)
}
