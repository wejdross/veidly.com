package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// VerifyEmail handles email verification via token
func VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required"})
		return
	}

	// Find the verification token in the database
	var tokenData struct {
		ID        int
		UserID    int
		ExpiresAt time.Time
	}

	err := db.QueryRow(`
		SELECT id, user_id, expires_at
		FROM email_verification_tokens
		WHERE token = ?
	`, token).Scan(&tokenData.ID, &tokenData.UserID, &tokenData.ExpiresAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification token"})
		return
	}
	if err != nil {
		log.Printf("Error querying verification token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if token has expired
	if time.Now().After(tokenData.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token has expired"})
		return
	}

	// Update user's email_verified status
	_, err = db.Exec(`UPDATE users SET email_verified = 1 WHERE id = ?`, tokenData.UserID)
	if err != nil {
		log.Printf("Error updating user email_verified status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	// Delete the used token
	_, err = db.Exec(`DELETE FROM email_verification_tokens WHERE id = ?`, tokenData.ID)
	if err != nil {
		log.Printf("Warning: Could not delete verification token: %v", err)
	}

	// Send welcome email if email service is available
	if emailService != nil {
		var user User
		err = db.QueryRow(`SELECT id, email, name FROM users WHERE id = ?`, tokenData.UserID).
			Scan(&user.ID, &user.Email, &user.Name)
		if err == nil {
			go emailService.SendWelcomeEmail(user.Email, user.Name)
		}
	}

	log.Printf("✓ Email verified for user ID: %d", tokenData.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// ResendVerificationEmail resends the verification email to a user
func ResendVerificationEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find user by email
	var user User
	err := db.QueryRow(`
		SELECT id, email, name, email_verified
		FROM users
		WHERE email = ?
	`, req.Email).Scan(&user.ID, &user.Email, &user.Name, &user.EmailVerified)

	if err == sql.ErrNoRows {
		// Don't reveal if email exists or not (security)
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a verification link has been sent"})
		return
	}
	if err != nil {
		log.Printf("Error querying user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if already verified
	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already verified"})
		return
	}

	// Check if email service is available
	if emailService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Email service is not configured"})
		return
	}

	// Delete any existing verification tokens for this user
	_, err = db.Exec(`DELETE FROM email_verification_tokens WHERE user_id = ?`, user.ID)
	if err != nil {
		log.Printf("Warning: Could not delete old verification tokens: %v", err)
	}

	// Generate new verification token
	token, err := generateEmailToken()
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verification token"})
		return
	}

	// Store token in database (expires in 24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = db.Exec(`
		INSERT INTO email_verification_tokens (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`, user.ID, token, expiresAt)
	if err != nil {
		log.Printf("Error storing verification token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verification token"})
		return
	}

	// Send verification email
	err = emailService.SendVerificationEmail(user.Email, user.Name, token)
	if err != nil {
		log.Printf("Error sending verification email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	log.Printf("✓ Verification email resent to: %s", user.Email)
	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent"})
}

// ForgotPassword handles password reset requests
func ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find user by email
	var user User
	err := db.QueryRow(`SELECT id, email, name FROM users WHERE email = ?`, req.Email).
		Scan(&user.ID, &user.Email, &user.Name)

	if err == sql.ErrNoRows {
		// Don't reveal if email exists or not (security)
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent"})
		return
	}
	if err != nil {
		log.Printf("Error querying user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if email service is available
	if emailService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Email service is not configured"})
		return
	}

	// Generate password reset token
	token, err := generateEmailToken()
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
		return
	}

	// Store token in database (expires in 1 hour)
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = db.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`, user.ID, token, expiresAt)
	if err != nil {
		log.Printf("Error storing reset token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset token"})
		return
	}

	// Send password reset email
	err = emailService.SendPasswordResetEmail(user.Email, user.Name, token)
	if err != nil {
		log.Printf("Error sending password reset email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email"})
		return
	}

	log.Printf("✓ Password reset email sent to: %s", user.Email)
	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent"})
}

// ResetPassword handles password reset via token
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find the reset token in the database
	var tokenData struct {
		ID        int
		UserID    int
		ExpiresAt time.Time
		Used      bool
	}

	err := db.QueryRow(`
		SELECT id, user_id, expires_at, used
		FROM password_reset_tokens
		WHERE token = ?
	`, req.Token).Scan(&tokenData.ID, &tokenData.UserID, &tokenData.ExpiresAt, &tokenData.Used)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}
	if err != nil {
		log.Printf("Error querying reset token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if token has already been used
	if tokenData.Used {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token has already been used"})
		return
	}

	// Check if token has expired
	if time.Now().After(tokenData.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token has expired"})
		return
	}

	// Hash the new password
	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Update user's password
	_, err = db.Exec(`UPDATE users SET password = ? WHERE id = ?`, hashedPassword, tokenData.UserID)
	if err != nil {
		log.Printf("Error updating password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	// Mark token as used
	_, err = db.Exec(`UPDATE password_reset_tokens SET used = 1 WHERE id = ?`, tokenData.ID)
	if err != nil {
		log.Printf("Warning: Could not mark reset token as used: %v", err)
	}

	log.Printf("✓ Password reset successful for user ID: %d", tokenData.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
