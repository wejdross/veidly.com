package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var emailService *EmailService

// isTestMode checks if we're running in test mode
func isTestMode() bool {
	// Check if any test flags are present
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

func parseDateTime(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			log.Printf("✓ Successfully parsed datetime: %s using format: %s", dateStr, format)
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", dateStr)
}

// migrateEventSlugs generates slugs for events that don't have them
func migrateEventSlugs() {
	rows, err := db.Query(`SELECT id, title FROM events WHERE slug IS NULL OR slug = ''`)
	if err != nil {
		log.Printf("⚠️  Error querying events for migration: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			log.Printf("⚠️  Error scanning event: %v", err)
			continue
		}

		slug := generateSlug(title)
		_, err := db.Exec(`UPDATE events SET slug = ? WHERE id = ?`, slug, id)
		if err != nil {
			log.Printf("⚠️  Error updating slug for event %d: %v", id, err)
		} else {
			count++
		}
	}

	if count > 0 {
		log.Printf("✓ Generated slugs for %d events", count)
	}
}

func initDB() {
	// SAFETY CHECK: Never initialize production DB during tests
	if isTestMode() {
		log.Println("⚠️  Test mode detected - skipping production database initialization")
		return
	}

	var err error
	// Enable WAL mode for better concurrency
	db, err = sql.Open("sqlite3", "./veidly.db?_journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("📦 Database connection established")
	log.Println("🔒 Production database: veidly.db")

	// Users table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		name TEXT NOT NULL,
		bio TEXT,
		threema TEXT,
		is_admin BOOLEAN DEFAULT 0,
		is_blocked BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Add new columns to existing users table (guarded)
	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='bio'`).Scan(&exists); err == nil && exists == 0 {
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN bio TEXT`); err != nil {
			log.Printf("⚠️  add bio failed: %v", err)
		}
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='threema'`).Scan(&exists); err == nil && exists == 0 {
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN threema TEXT`); err != nil {
			log.Printf("⚠️  add threema failed: %v", err)
		}
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='languages'`).Scan(&exists); err == nil && exists == 0 {
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN languages TEXT`); err != nil {
			log.Printf("⚠️  add languages failed: %v", err)
		}
	}

	// Optional data migration from telegram -> threema if telegram column exists
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='telegram'`).Scan(&exists); err == nil && exists == 1 {
		if _, err := db.Exec(`UPDATE users SET threema = COALESCE(threema, telegram) WHERE telegram IS NOT NULL AND telegram != ''`); err != nil {
			log.Printf("⚠️  migrate telegram->threema failed: %v", err)
		}
		// Do NOT drop columns here; SQLite drop column is version-dependent and risky.
	}

	// Events table (updated with user_id and category)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		category TEXT NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		creator_name TEXT NOT NULL,
		max_participants INTEGER,
		gender_restriction TEXT DEFAULT 'any',
		age_min INTEGER DEFAULT 0,
		age_max INTEGER DEFAULT 99,
		smoking_allowed BOOLEAN DEFAULT 0,
		alcohol_allowed BOOLEAN DEFAULT 0,
		slug TEXT UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Add slug column to existing events table (migration)
	// Check if slug column exists
	var slugExists int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='slug'`).Scan(&slugExists)
	if slugExists == 0 {
		log.Println("📝 Adding slug column to events table...")
		_, err = db.Exec(`ALTER TABLE events ADD COLUMN slug TEXT`)
		if err != nil {
			log.Printf("⚠️  Warning: Could not add slug column: %v", err)
		} else {
			log.Println("✓ Slug column added successfully")
			// Generate slugs for existing events
			migrateEventSlugs()
		}
	} else {
		// Check if there are events without slugs
		var eventsWithoutSlug int
		db.QueryRow(`SELECT COUNT(*) FROM events WHERE slug IS NULL OR slug = ''`).Scan(&eventsWithoutSlug)
		if eventsWithoutSlug > 0 {
			log.Printf("📝 Found %d events without slugs, generating...", eventsWithoutSlug)
			migrateEventSlugs()
		}
	}

	// Add event_languages column to events table (migration)
	var eventLanguagesExists int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='event_languages'`).Scan(&eventLanguagesExists)
	if eventLanguagesExists == 0 {
		log.Println("📝 Adding event_languages column to events table...")
		_, err = db.Exec(`ALTER TABLE events ADD COLUMN event_languages TEXT`)
		if err != nil {
			log.Printf("⚠️  Warning: Could not add event_languages column: %v", err)
		} else {
			log.Println("✓ event_languages column added successfully")
		}
	}

	// Add email_verified column to users table (migration)
	var emailVerifiedExists int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='email_verified'`).Scan(&emailVerifiedExists)
	if emailVerifiedExists == 0 {
		log.Println("📝 Adding email_verified column to users table...")
		_, err = db.Exec(`ALTER TABLE users ADD COLUMN email_verified INTEGER DEFAULT 0`)
		if err != nil {
			log.Printf("⚠️  Warning: Could not add email_verified column: %v", err)
		} else {
			log.Println("✓ email_verified column added successfully")
			// Auto-verify existing users (backward compatibility)
			_, err = db.Exec(`UPDATE users SET email_verified = 1 WHERE email_verified = 0`)
			if err != nil {
				log.Printf("⚠️  Warning: Could not update existing users: %v", err)
			} else {
				log.Println("✓ Existing users auto-verified for backward compatibility")
			}
		}
	}

	// Add privacy control columns to events table (migration)
	privacyColumns := map[string]string{
		"hide_organizer_until_joined":    "INTEGER DEFAULT 0",
		"hide_participants_until_joined": "INTEGER DEFAULT 1",
		"require_verified_to_join":       "INTEGER DEFAULT 0",
		"require_verified_to_view":       "INTEGER DEFAULT 0",
	}

	// Whitelist of allowed column names to prevent SQL injection
	allowedColumns := map[string]bool{
		"hide_organizer_until_joined":    true,
		"hide_participants_until_joined": true,
		"require_verified_to_join":       true,
		"require_verified_to_view":       true,
	}

	for columnName, columnDef := range privacyColumns {
		// Validate column name is in whitelist
		if !allowedColumns[columnName] {
			log.Printf("⚠️  Skipping invalid column name: %s", columnName)
			continue
		}

		var colExists int
		// Use parameterized query for column name check
		err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('events') WHERE name = ?`, columnName).Scan(&colExists)
		if err != nil {
			log.Printf("⚠️  Warning: Could not check %s column: %v", columnName, err)
			continue
		}

		if colExists == 0 {
			log.Printf("📝 Adding %s column to events table...", columnName)
			// Column name is now validated against whitelist, safe to use in query
			_, err = db.Exec(fmt.Sprintf(`ALTER TABLE events ADD COLUMN %s %s`, columnName, columnDef))
			if err != nil {
				log.Printf("⚠️  Warning: Could not add %s column: %v", columnName, err)
			} else {
				log.Printf("✓ %s column added successfully", columnName)
			}
		}
	}

	// Remove drugs_allowed column if it exists (migration)
	var drugsColExists int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='drugs_allowed'`).Scan(&drugsColExists)
	if drugsColExists > 0 {
		log.Printf("📝 Removing drugs_allowed column from events table...")
		// SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
		// However, since this is a minor column removal and doesn't affect functionality,
		// we'll just leave it in existing databases and it won't be used
		log.Printf("ℹ️  Note: drugs_allowed column exists but will be ignored (SQLite limitation)")
	}

	// Note: creator_contact column may exist in legacy databases but is no longer used
	// Privacy improvement: contact information is now only accessible via comments to participants
	var creatorContactExists int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='creator_contact'`).Scan(&creatorContactExists)
	if creatorContactExists > 0 {
		log.Printf("ℹ️  Note: creator_contact column exists but is no longer used (privacy improvement)")
	}

	// Email verification tokens table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS email_verification_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Password reset tokens table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		used INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Event participants table (tracks who's attending events)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS event_participants (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
		UNIQUE(event_id, user_id)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// User blocking table (bidirectional blocking for privacy)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS user_blocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		blocker_id INTEGER NOT NULL,
		blocked_id INTEGER NOT NULL,
		reason TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (blocker_id) REFERENCES users (id) ON DELETE CASCADE,
		FOREIGN KEY (blocked_id) REFERENCES users (id) ON DELETE CASCADE,
		UNIQUE(blocker_id, blocked_id)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Create indexes for user_blocks
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_blocks_blocker ON user_blocks(blocker_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_blocks_blocked ON user_blocks(blocked_id)`)

	// Event comments table (participant-only communication)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS event_comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		comment TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME,
		is_deleted BOOLEAN DEFAULT 0,
		FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Create indexes for event_comments
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_comments_event ON event_comments(event_id, created_at)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_comments_user ON event_comments(user_id)`)

	// Event reports table (moderation system)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS event_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER NOT NULL,
		reporter_id INTEGER NOT NULL,
		reason TEXT NOT NULL,
		description TEXT,
		status TEXT DEFAULT 'pending',
		reviewed_by INTEGER,
		reviewed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
		FOREIGN KEY (reporter_id) REFERENCES users (id) ON DELETE CASCADE,
		FOREIGN KEY (reviewed_by) REFERENCES users (id)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_reports_status ON event_reports(status, created_at)`)

	// Comment reports table (moderation system)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS comment_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		comment_id INTEGER NOT NULL,
		reporter_id INTEGER NOT NULL,
		reason TEXT NOT NULL,
		description TEXT,
		status TEXT DEFAULT 'pending',
		reviewed_by INTEGER,
		reviewed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (comment_id) REFERENCES event_comments (id) ON DELETE CASCADE,
		FOREIGN KEY (reporter_id) REFERENCES users (id) ON DELETE CASCADE,
		FOREIGN KEY (reviewed_by) REFERENCES users (id)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Add comments_enabled column to events table (migration)
	var commentsEnabledExists int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='comments_enabled'`).Scan(&commentsEnabledExists)
	if commentsEnabledExists == 0 {
		log.Println("📝 Adding comments_enabled column to events table...")
		_, err = db.Exec(`ALTER TABLE events ADD COLUMN comments_enabled BOOLEAN DEFAULT 1`)
		if err != nil {
			log.Printf("⚠️  Warning: Could not add comments_enabled column: %v", err)
		} else {
			log.Println("✓ comments_enabled column added successfully")
		}
	}

	// Create default admin user with secure password
	var adminCount int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = 1").Scan(&adminCount)

	if adminCount == 0 {
		adminPass := os.Getenv("ADMIN_PASSWORD")
		if adminPass == "" {
			// Generate random password only in development
			if os.Getenv("ENVIRONMENT") == "development" {
				adminPass = generateRandomString(16)
				log.Printf("⚠️  IMPORTANT: Generated admin password: %s", adminPass)
				log.Println("⚠️  SAVE THIS PASSWORD - Set ADMIN_PASSWORD env var for custom password")
			} else {
				log.Fatal("ADMIN_PASSWORD environment variable must be set in production")
			}
		}
		hashedPassword, err := hashPassword(adminPass)
		if err != nil {
			log.Fatalf("Failed to hash admin password: %v", err)
		}
		result, err := db.Exec(`INSERT INTO users (email, password, name, is_admin) VALUES (?, ?, 'Admin', 1)`,
			"admin@veidly.com", hashedPassword)
		if err != nil {
			log.Printf("⚠️  Admin user creation failed: %v", err)
		} else {
			id, err := result.LastInsertId()
			if err != nil {
				log.Printf("⚠️  Admin user created but failed to get ID: %v", err)
			} else {
				log.Printf("✓ Admin user created with ID: %d (email: admin@veidly.com)", id)
			}
		}
	}

	// Create indexes for better query performance
	log.Println("📊 Creating database indexes...")
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_category ON events(category)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_participants_event_id ON event_participants(event_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_participants_user_id ON event_participants(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_events_slug ON events(slug)`)
	log.Println("✓ Indexes ready")

	log.Println("✓ Database schema ready")
}

func main() {
	log.Println("🚀 Starting Veidly Server...")
	// Initialize JWT secret
	if err := initJWTFromEnv(); err != nil {
		log.Fatalf("JWT init error: %v", err)
	}
	initDB()
	if db != nil {
		defer db.Close()
	}

	// Initialize email service
	emailService = NewEmailService()

	// Set Gin mode based on environment
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add custom middleware
	router.Use(RequestIDMiddleware())
	router.Use(LoggerMiddleware())
	router.Use(gin.Recovery())
	router.Use(ErrorHandlerMiddleware())
	router.Use(SecurityHeadersMiddleware())
	router.Use(RequestSizeLimitMiddleware(5 * 1024 * 1024)) // 5MB limit

	// CORS middleware (origins from env CORS_ORIGINS, comma-separated)
	originsEnv := os.Getenv("CORS_ORIGINS")
	var allowedOrigins []string

	if strings.TrimSpace(originsEnv) == "" {
		// Development defaults
		if os.Getenv("ENVIRONMENT") != "production" {
			allowedOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
			log.Println("⚠️  Using default CORS origins (development mode)")
		} else {
			log.Fatal("CORS_ORIGINS must be set in production")
		}
	} else {
		// Validate and parse CORS origins
		if strings.Contains(originsEnv, "*") {
			log.Fatal("Wildcard CORS origins (*) are not allowed for security")
		}
		parts := strings.Split(originsEnv, ",")
		for _, p := range parts {
			v := strings.TrimSpace(p)
			if v != "" {
				allowedOrigins = append(allowedOrigins, v)
			}
		}
		log.Printf("✓ CORS origins: %v", allowedOrigins)
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rate limiters for different endpoints (increased for testing/seeding)
	// Store limiter instances for graceful shutdown
	authLimiterInstance, authLimiter := RateLimitMiddleware(20, time.Minute)             // 20 requests per minute for auth (5 in production)
	apiLimiterInstance, apiLimiter := RateLimitMiddleware(200, time.Minute)              // 200 requests per minute for API (100 in production)
	searchLimiterInstance, searchLimiter := RateLimitMiddleware(50, time.Minute)         // 50 searches per minute (30 in production)
	createEventLimiterInstance, createEventLimiter := RateLimitMiddleware(100, time.Hour) // 100 events per hour (10 in production)

	// Collect all limiters for shutdown
	rateLimiters := []*rateLimiter{authLimiterInstance, apiLimiterInstance, searchLimiterInstance, createEventLimiterInstance}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now()})
	})

	// Public routes with rate limiting
	router.POST("/api/auth/register", authLimiter, register)
	router.POST("/api/auth/login", authLimiter, login)
	router.POST("/api/auth/logout", authLimiter, logout)                           // Logout (clears httpOnly cookie)
	router.GET("/api/auth/verify-email", apiLimiter, VerifyEmail)                  // Email verification
	router.POST("/api/auth/resend-verification", authLimiter, ResendVerificationEmail) // Resend verification
	router.POST("/api/auth/forgot-password", authLimiter, ForgotPassword)          // Password reset request
	router.POST("/api/auth/reset-password", authLimiter, ResetPassword)            // Password reset
	router.GET("/api/events", apiLimiter, optionalAuthMiddleware(), getEvents)
	router.GET("/api/events/:id", apiLimiter, optionalAuthMiddleware(), getEvent)
	router.GET("/api/events/:id/participants", apiLimiter, optionalAuthMiddleware(), getEventParticipants)
	router.GET("/api/public/events/:slug", apiLimiter, optionalAuthMiddleware(), getPublicEvent) // Public event access by slug
	router.GET("/api/public/events/:slug/ics", apiLimiter, downloadEventICS)                  // Download ICS calendar file
	router.GET("/api/search/places", searchLimiter, searchPlaces)
	router.GET("/api/categories", getCategories)

	// Protected routes (require authentication)
	protected := router.Group("/api")
	protected.Use(authMiddleware())
	{
		protected.POST("/events", createEventLimiter, createEvent)
		protected.PUT("/events/:id", updateEvent)
		protected.DELETE("/events/:id", deleteEvent)
		protected.POST("/events/:id/join", joinEvent)
		protected.DELETE("/events/:id/leave", leaveEvent)
		protected.GET("/auth/me", getCurrentUser)
		protected.GET("/profile", getOwnProfile)
		protected.PUT("/profile", updateProfile)
		protected.GET("/profile/:id", getUserProfile)

		// Blocking routes
		protected.POST("/users/:id/block", blockUser)
		protected.DELETE("/users/:id/block", unblockUser)
		protected.GET("/blocks", getBlockedUsers)

		// Comment routes
		protected.GET("/events/:id/comments", getEventComments)
		protected.POST("/events/:id/comments", createEventComment)
		protected.PUT("/comments/:id", updateEventComment)
		protected.DELETE("/comments/:id", deleteEventComment)
	}

	// Admin routes
	admin := router.Group("/api/admin")
	admin.Use(authMiddleware(), adminMiddleware())
	{
		admin.GET("/users", adminGetUsers)
		admin.PUT("/users/:id/block", adminBlockUser)
		admin.PUT("/users/:id/unblock", adminUnblockUser)
		admin.PUT("/users/:id/verify-email", adminVerifyUserEmail)
		admin.GET("/events", adminGetAllEvents)
		admin.DELETE("/events/:id", adminDeleteEvent)
		admin.PUT("/events/:id", adminUpdateEvent)
	}

	port := os.Getenv("PORT")
	if strings.TrimSpace(port) == "" {
		port = os.Getenv("SERVER_PORT")
	}
	if strings.TrimSpace(port) == "" {
		port = "8080"
	}

	// TLS/HTTPS support
	useTLS := os.Getenv("USE_TLS") == "true"

	// Create HTTP server with proper configuration
	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		if useTLS {
			certFile := os.Getenv("TLS_CERT")
			keyFile := os.Getenv("TLS_KEY")

			if certFile == "" || keyFile == "" {
				log.Fatal("TLS_CERT and TLS_KEY must be set when USE_TLS=true")
			}

			log.Printf("✓ Server running on https://localhost:%s (TLS enabled)", port)
			log.Println("✓ API endpoints ready")
			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		} else {
			// When behind a reverse proxy (like nginx) that handles TLS termination,
			// it's acceptable to run the backend on HTTP locally
			log.Printf("✓ Server running on http://localhost:%s", port)
			if os.Getenv("ENVIRONMENT") == "production" {
				log.Println("⚠️  Running without TLS - ensure reverse proxy (nginx) handles HTTPS")
			}
			log.Println("✓ API endpoints ready")
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// SIGINT handles Ctrl+C, SIGTERM handles kill command
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Shutdown all rate limiter cleanup goroutines
	for _, limiter := range rateLimiters {
		limiter.Shutdown()
	}

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("⚠️  Server forced to shutdown:", err)
	}

	log.Println("✓ Server exited gracefully")
}
