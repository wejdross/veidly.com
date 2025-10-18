package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

// EmailService handles all email-related operations
type EmailService struct {
	mg     *mailgun.MailgunImpl
	domain string
	from   string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	domain := os.Getenv("MAILGUN_DOMAIN")
	apiKey := os.Getenv("MAILGUN_API_KEY")
	from := os.Getenv("MAILGUN_FROM_EMAIL")

	if domain == "" || apiKey == "" {
		log.Println("⚠️  Mailgun not configured - email features disabled")
		return nil
	}

	mg := mailgun.NewMailgun(domain, apiKey)

	// Use EU API endpoint for EU domains (veidly.com)
	// Change this to https://api.mailgun.net for US domains
	// Note: Do NOT include /v3 - the library adds it automatically
	mg.SetAPIBase("https://api.eu.mailgun.net")

	log.Printf("✓ Email service initialized for domain: %s (EU endpoint)", domain)
	return &EmailService{
		mg:     mg,
		domain: domain,
		from:   from,
	}
}

// generateEmailToken generates a secure random token for email verification
func generateEmailToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SendVerificationEmail sends an email verification link
func (s *EmailService) SendVerificationEmail(email, name, token string) error {
	if s == nil {
		log.Println("⚠️  Email service not available - skipping verification email")
		return nil
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}

	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)

	subject := "Verify your Veidly account"
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 15px 30px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; text-decoration: none; border-radius: 50px; font-weight: bold; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to Veidly!</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>Thank you for signing up! Please verify your email address to start creating and joining events.</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Verify Email Address</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #667eea;">%s</p>
            <p><strong>This link will expire in 24 hours.</strong></p>
            <p>If you didn't create an account, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>© 2025 Veidly - Connect and meet new people</p>
        </div>
    </div>
</body>
</html>
`, name, verificationLink, verificationLink)

	textBody := fmt.Sprintf(`
Hi %s,

Thank you for signing up for Veidly!

Please verify your email address by clicking the link below:
%s

This link will expire in 24 hours.

If you didn't create an account, you can safely ignore this email.

© 2025 Veidly - Connect and meet new people
`, name, verificationLink)

	message := s.mg.NewMessage(s.from, subject, textBody, email)
	message.SetHtml(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, _, err := s.mg.Send(ctx, message)
	if err != nil {
		log.Printf("❌ Failed to send verification email to %s: %v", email, err)
		return err
	}

	log.Printf("✓ Verification email sent to %s", email)
	return nil
}

// SendPasswordResetEmail sends a password reset link
func (s *EmailService) SendPasswordResetEmail(email, name, token string) error {
	if s == nil {
		log.Println("⚠️  Email service not available - skipping password reset email")
		return nil
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

	subject := "Reset your Veidly password"
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 15px 30px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; text-decoration: none; border-radius: 50px; font-weight: bold; margin: 20px 0; }
        .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #667eea;">%s</p>
            <div class="warning">
                <p><strong>⚠️ Security Notice:</strong></p>
                <p>This link will expire in 1 hour. If you didn't request a password reset, please ignore this email and your password will remain unchanged.</p>
            </div>
        </div>
        <div class="footer">
            <p>© 2025 Veidly - Connect and meet new people</p>
        </div>
    </div>
</body>
</html>
`, name, resetLink, resetLink)

	textBody := fmt.Sprintf(`
Hi %s,

We received a request to reset your password for your Veidly account.

Click the link below to reset your password:
%s

This link will expire in 1 hour.

If you didn't request a password reset, please ignore this email and your password will remain unchanged.

© 2025 Veidly - Connect and meet new people
`, name, resetLink)

	message := s.mg.NewMessage(s.from, subject, textBody, email)
	message.SetHtml(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, _, err := s.mg.Send(ctx, message)
	if err != nil {
		log.Printf("❌ Failed to send password reset email to %s: %v", email, err)
		return err
	}

	log.Printf("✓ Password reset email sent to %s", email)
	return nil
}

// SendWelcomeEmail sends a welcome email after verification
func (s *EmailService) SendWelcomeEmail(email, name string) error {
	if s == nil {
		return nil
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}

	subject := "Welcome to Veidly - Let's get started!"
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 15px 30px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; text-decoration: none; border-radius: 50px; font-weight: bold; margin: 20px 0; }
        .features { background: white; padding: 20px; border-radius: 10px; margin: 20px 0; }
        .feature { margin: 15px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 You're all set!</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>Your email has been verified and your account is now active! You can now:</p>
            <div class="features">
                <div class="feature">✅ Create events and meet new people</div>
                <div class="feature">🔍 Browse and join events near you</div>
                <div class="feature">💬 Connect with other members</div>
                <div class="feature">🌍 Filter events by language, interests, and more</div>
            </div>
            <p style="text-align: center;">
                <a href="%s" class="button">Start Exploring Events</a>
            </p>
            <p>Have fun connecting with people!</p>
        </div>
        <div class="footer">
            <p>© 2025 Veidly - Connect and meet new people</p>
        </div>
    </div>
</body>
</html>
`, name, baseURL)

	textBody := fmt.Sprintf(`
Hi %s,

Your email has been verified and your account is now active!

You can now:
- Create events and meet new people
- Browse and join events near you
- Connect with other members
- Filter events by language, interests, and more

Start exploring: %s

Have fun connecting with people!

© 2025 Veidly - Connect and meet new people
`, name, baseURL)

	message := s.mg.NewMessage(s.from, subject, textBody, email)
	message.SetHtml(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, _, err := s.mg.Send(ctx, message)
	if err != nil {
		log.Printf("❌ Failed to send welcome email to %s: %v", email, err)
		return err
	}

	log.Printf("✓ Welcome email sent to %s", email)
	return nil
}
