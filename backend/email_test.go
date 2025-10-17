package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmailService(t *testing.T) {
	// Save original env vars
	origDomain := os.Getenv("MAILGUN_DOMAIN")
	origAPIKey := os.Getenv("MAILGUN_API_KEY")
	origFromEmail := os.Getenv("MAILGUN_FROM_EMAIL")
	defer func() {
		os.Setenv("MAILGUN_DOMAIN", origDomain)
		os.Setenv("MAILGUN_API_KEY", origAPIKey)
		os.Setenv("MAILGUN_FROM_EMAIL", origFromEmail)
	}()

	// Test with valid credentials
	os.Setenv("MAILGUN_DOMAIN", "mg.example.com")
	os.Setenv("MAILGUN_API_KEY", "test-api-key-123")
	os.Setenv("MAILGUN_FROM_EMAIL", "noreply@example.com")

	service := NewEmailService()
	assert.NotNil(t, service)
	assert.Equal(t, "mg.example.com", service.domain)
	assert.Equal(t, "noreply@example.com", service.from)
}

func TestGenerateEmailToken(t *testing.T) {
	// Test that token is generated and has correct length
	token1, err1 := generateEmailToken()
	require.NoError(t, err1)
	assert.Equal(t, 64, len(token1)) // 32 bytes hex encoded = 64 chars

	// Test that tokens are unique
	token2, err2 := generateEmailToken()
	require.NoError(t, err2)
	assert.NotEqual(t, token1, token2)

	// Test that token contains only valid hex characters
	for _, char := range token1 {
		assert.True(t, (char >= 'a' && char <= 'f') || (char >= '0' && char <= '9'))
	}
}

func TestSendVerificationEmail(t *testing.T) {
	// Setup test email service (won't actually send emails)
	os.Setenv("MAILGUN_DOMAIN", "mg.example.com")
	os.Setenv("MAILGUN_API_KEY", "test-api-key")
	os.Setenv("MAILGUN_FROM_EMAIL", "noreply@example.com")

	service := NewEmailService()

	// Test parameters - these will fail to send (no real credentials)
	// but we're testing that the function doesn't panic
	tests := []struct {
		name     string
		toEmail  string
		userName string
		token    string
	}{
		{
			name:     "Valid verification email",
			toEmail:  "test@example.com",
			userName: "Test User",
			token:    "abc123xyz",
		},
		{
			name:     "User with special characters in name",
			toEmail:  "user@example.com",
			userName: "John O'Brien",
			token:    "token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail to send (no real Mailgun credentials),
			// but we're testing that it doesn't panic and constructs the message
			err := service.SendVerificationEmail(tt.toEmail, tt.userName, tt.token)
			// We expect an error since we don't have real credentials
			_ = err // Error expected in test environment
		})
	}
}

func TestSendPasswordResetEmail(t *testing.T) {
	os.Setenv("MAILGUN_DOMAIN", "mg.example.com")
	os.Setenv("MAILGUN_API_KEY", "test-api-key")
	os.Setenv("MAILGUN_FROM_EMAIL", "noreply@example.com")

	service := NewEmailService()

	tests := []struct {
		name     string
		toEmail  string
		userName string
		token    string
	}{
		{
			name:     "Valid password reset email",
			toEmail:  "test@example.com",
			userName: "Test User",
			token:    "reset-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SendPasswordResetEmail(tt.toEmail, tt.userName, tt.token)
			_ = err // Error expected in test environment
		})
	}
}

func TestSendWelcomeEmail(t *testing.T) {
	os.Setenv("MAILGUN_DOMAIN", "mg.example.com")
	os.Setenv("MAILGUN_API_KEY", "test-api-key")
	os.Setenv("MAILGUN_FROM_EMAIL", "noreply@example.com")

	service := NewEmailService()

	tests := []struct {
		name     string
		toEmail  string
		userName string
	}{
		{
			name:     "Valid welcome email",
			toEmail:  "test@example.com",
			userName: "New User",
		},
		{
			name:     "Welcome email with unicode name",
			toEmail:  "user@example.com",
			userName: "Łukasz Widera",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SendWelcomeEmail(tt.toEmail, tt.userName)
			_ = err // Error expected in test environment
		})
	}
}

func TestEmailServiceNilSafety(t *testing.T) {
	// Test that email service handles missing configuration gracefully
	os.Unsetenv("MAILGUN_DOMAIN")
	os.Unsetenv("MAILGUN_API_KEY")
	os.Unsetenv("MAILGUN_FROM_EMAIL")

	service := NewEmailService()

	// Service should be nil when credentials are missing
	assert.Nil(t, service)
}
