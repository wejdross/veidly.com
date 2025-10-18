package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)


// Auth handlers
func register(c *gin.Context) {
	log.Println("📝 POST /api/auth/register - New user registration")

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Invalid registration data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("❌ Password hashing failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	// Insert user (email_verified defaults to false/0)
	result, err := db.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, 0)
	`, req.Email, hashedPassword, req.Name)

	if err != nil {
		log.Printf("❌ User registration failed: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("❌ Failed to get user ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	// Create user object for response
	user := User{
		ID:            int(id),
		Email:         req.Email,
		Name:          req.Name,
		IsAdmin:       false,
		IsBlocked:     false,
		EmailVerified: false,
		CreatedAt:     time.Now(),
	}

	// Send verification email if email service is configured
	if emailService != nil {
		// Generate verification token
		verificationToken, err := generateEmailToken()
		if err != nil {
			log.Printf("⚠️  Warning: Could not generate verification token: %v", err)
		} else {
			// Store token in database (expires in 24 hours)
			expiresAt := time.Now().Add(24 * time.Hour)
			_, err = db.Exec(`
				INSERT INTO email_verification_tokens (user_id, token, expires_at)
				VALUES (?, ?, ?)
			`, user.ID, verificationToken, expiresAt)
			if err != nil {
				log.Printf("⚠️  Warning: Could not store verification token: %v", err)
			} else {
				// Send verification email asynchronously
				go func() {
					err := emailService.SendVerificationEmail(user.Email, user.Name, verificationToken)
					if err != nil {
						log.Printf("⚠️  Warning: Could not send verification email to %s: %v", user.Email, err)
					} else {
						log.Printf("📧 Verification email sent to: %s", user.Email)
					}
				}()
			}
		}
	} else {
		log.Println("⚠️  Email service not configured - skipping verification email")
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		log.Printf("❌ Token generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Printf("✅ User registered successfully: %s (ID: %d)", user.Email, user.ID)
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}

func login(c *gin.Context) {
	log.Println("🔐 POST /api/auth/login - User login attempt")

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Invalid login data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login data"})
		return
	}

	var user User
	var hashedPassword string
	var bio, languages sql.NullString
	err := db.QueryRow(`
		SELECT id, email, password, name, bio, languages, is_admin, is_blocked, email_verified, created_at
		FROM users WHERE email = ?
	`, req.Email).Scan(&user.ID, &user.Email, &hashedPassword, &user.Name, &bio, &languages,
		&user.IsAdmin, &user.IsBlocked, &user.EmailVerified, &user.CreatedAt)

	// Convert NullString to string
	if bio.Valid {
		user.Bio = bio.String
	}
	if languages.Valid {
		user.Languages = languages.String
	}

	if err == sql.ErrNoRows {
		log.Printf("❌ Login failed: User not found - %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		log.Printf("❌ Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	if user.IsBlocked {
		log.Printf("❌ Login denied: User is blocked - %s", req.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is blocked"})
		return
	}

	if !checkPasswordHash(req.Password, hashedPassword) {
		log.Printf("❌ Login failed: Invalid password - %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := generateToken(user)
	if err != nil {
		log.Printf("❌ Token generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Printf("✅ User logged in successfully: %s", user.Email)
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func logout(c *gin.Context) {
	log.Println("🚪 POST /api/auth/logout - User logout")
	log.Println("✅ User logged out successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func getCurrentUser(c *gin.Context) {
	userID := c.GetInt("user_id")

	var user User
	var bio, languages sql.NullString
	err := db.QueryRow(`
		SELECT id, email, name, bio, languages, is_admin, is_blocked, email_verified, created_at
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Email, &user.Name, &bio, &languages, &user.IsAdmin, &user.IsBlocked, &user.EmailVerified, &user.CreatedAt)

	// Convert NullString to string
	if bio.Valid {
		user.Bio = bio.String
	}
	if languages.Valid {
		user.Languages = languages.String
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return user wrapped in object for consistency with login/register
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// Event handlers
func getEvents(c *gin.Context) {
	log.Println("📋 GET /api/events - Fetching all upcoming events")

	// Get query parameters
	category := c.Query("category")
	keyword := c.Query("keyword")
	location := c.Query("location")
	languages := c.Query("languages")
	smokingAllowed := c.Query("smoking")
	alcoholAllowed := c.Query("alcohol")
	gender := c.Query("gender")
	ageMin := c.Query("age_min")
	ageMax := c.Query("age_max")

	// Get viewer info for privacy filtering
	viewerUserID, _ := c.Get("user_id")
	viewerIsAdmin, _ := c.Get("is_admin")
	viewerIsVerified, _ := c.Get("email_verified")

	userID := 0
	isAdmin := false
	isVerified := false

	if viewerUserID != nil {
		userID, _ = viewerUserID.(int)
	}
	if viewerIsAdmin != nil {
		isAdmin, _ = viewerIsAdmin.(bool)
	}
	if viewerIsVerified != nil {
		isVerified, _ = viewerIsVerified.(bool)
	}

	args := []interface{}{}

	query := `
		SELECT e.id, e.user_id, e.title, e.description, e.category, e.latitude, e.longitude,
		       e.start_time, e.end_time, e.creator_name, e.max_participants,
		       e.gender_restriction, e.age_min, e.age_max,
		       e.smoking_allowed, e.alcohol_allowed, e.event_languages, e.slug, e.created_at,
		       e.hide_organizer_until_joined, e.hide_participants_until_joined,
		       e.require_verified_to_join, e.require_verified_to_view,
		       u.email, u.languages as creator_languages,
		       (SELECT COUNT(*) FROM event_participants WHERE event_id = e.id) as participant_count
	`

	// If user is authenticated, check if they're a participant
	if userID > 0 {
		query += `, (SELECT COUNT(*) > 0 FROM event_participants WHERE event_id = e.id AND user_id = ?) as is_participant`
		args = append(args, userID)
	} else {
		query += `, 0 as is_participant`
	}

	query += `
		FROM events e
		LEFT JOIN users u ON e.user_id = u.id
		WHERE e.start_time >= datetime('now')
		AND e.start_time <= datetime('now', '+1 month')
	`

	// Category filter
	if category != "" {
		query += " AND e.category = ?"
		args = append(args, category)
	}

	// Keyword search (title or description)
	if keyword != "" {
		query += " AND (e.title LIKE ? OR e.description LIKE ?)"
		likeKeyword := "%" + keyword + "%"
		args = append(args, likeKeyword, likeKeyword)
	}

	// Location search (title/description contains location)
	if location != "" {
		query += " AND (e.title LIKE ? OR e.description LIKE ?)"
		likeLocation := "%" + location + "%"
		args = append(args, likeLocation, likeLocation)
	}

	// Language filter (search in event_languages, supports comma-separated list)
	if languages != "" {
		// Whitelist of valid language codes to prevent SQL injection via LIKE wildcards
		validLanguageCodes := map[string]bool{
			"bg": true, "hr": true, "cs": true, "da": true, "nl": true, "en": true,
			"et": true, "fi": true, "fr": true, "de": true, "el": true, "hu": true,
			"ga": true, "it": true, "lv": true, "lt": true, "mt": true, "pl": true,
			"pt": true, "ro": true, "sk": true, "sl": true, "es": true, "sv": true,
			"rm": true, "tr": true, "ar": true, "ru": true, "uk": true, "zh": true,
		}

		// Split comma-separated language codes and search for any match
		langCodes := strings.Split(languages, ",")
		if len(langCodes) > 0 {
			langConditions := []string{}
			for _, code := range langCodes {
				code = strings.TrimSpace(code)
				// Validate language code against whitelist
				if code != "" && validLanguageCodes[code] {
					langConditions = append(langConditions, "e.event_languages LIKE ?")
					args = append(args, "%"+code+"%")
				}
			}
			if len(langConditions) > 0 {
				query += " AND (" + strings.Join(langConditions, " OR ") + ")"
			}
		}
	}

	// Smoking filter
	if smokingAllowed == "true" {
		query += " AND e.smoking_allowed = 1"
	} else if smokingAllowed == "false" {
		query += " AND e.smoking_allowed = 0"
	}

	// Alcohol filter
	if alcoholAllowed == "true" {
		query += " AND e.alcohol_allowed = 1"
	} else if alcoholAllowed == "false" {
		query += " AND e.alcohol_allowed = 0"
	}

	// Gender filter
	if gender != "" && gender != "any" {
		query += " AND (e.gender_restriction = ? OR e.gender_restriction = 'any')"
		args = append(args, gender)
	}

	// Age filters
	if ageMin != "" {
		query += " AND e.age_max >= ?"
		args = append(args, ageMin)
	}
	if ageMax != "" {
		query += " AND e.age_min <= ?"
		args = append(args, ageMax)
	}

	query += " ORDER BY e.start_time ASC LIMIT 100"

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("❌ Error querying events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var startTime, endTime, genderRestriction, eventLanguages, creatorLanguages, slug, userEmail sql.NullString
		var maxParticipants sql.NullInt64
		var createdAt time.Time
		var isParticipant bool
		err := rows.Scan(
			&e.ID, &e.UserID, &e.Title, &e.Description, &e.Category, &e.Latitude, &e.Longitude,
			&startTime, &endTime, &e.CreatorName,
			&maxParticipants, &genderRestriction, &e.AgeMin, &e.AgeMax,
			&e.SmokingAllowed, &e.AlcoholAllowed, &eventLanguages, &slug, &createdAt,
			&e.HideOrganizerUntilJoined, &e.HideParticipantsUntilJoined,
			&e.RequireVerifiedToJoin, &e.RequireVerifiedToView,
			&userEmail, &creatorLanguages, &e.ParticipantCount, &isParticipant,
		)
		if err != nil {
			log.Printf("❌ Error scanning event: %v", err)
			continue
		}
		if startTime.Valid {
			e.StartTime = startTime.String
		}
		if endTime.Valid {
			e.EndTime = endTime.String
		}
		if maxParticipants.Valid {
			e.MaxParticipants = int(maxParticipants.Int64)
		}
		if genderRestriction.Valid {
			e.GenderRestriction = genderRestriction.String
		} else {
			e.GenderRestriction = "any"
		}
		if eventLanguages.Valid {
			e.EventLanguages = eventLanguages.String
		}
		if creatorLanguages.Valid {
			e.CreatorLanguages = creatorLanguages.String
		}
		if slug.Valid {
			e.Slug = slug.String
		}
		if userEmail.Valid {
			e.UserEmail = userEmail.String
		}
		e.CreatedAt = createdAt
		e.IsParticipant = isParticipant

		// Check if event can be viewed
		if errMsg := CheckEventViewPermission(&e, userID, isVerified, isAdmin); errMsg != "" {
			// Skip events that require verification
			continue
		}

		// Apply privacy filters
		ApplyPrivacyFilters(&e, userID, isVerified, isAdmin)

		events = append(events, e)
	}

	log.Printf("✓ Found %d events", len(events))
	c.JSON(http.StatusOK, events)
}

func getEvent(c *gin.Context) {
	id := c.Param("id")
	log.Printf("📖 GET /api/events/%s - Fetching single event", id)

	var e Event
	var startTime, endTime, genderRestriction, eventLanguages, slug sql.NullString
	var maxParticipants sql.NullInt64
	var createdAt time.Time
	err := db.QueryRow(`
		SELECT e.id, e.user_id, e.title, e.description, e.category, e.latitude, e.longitude,
		       e.start_time, e.end_time, e.creator_name, e.max_participants,
		       e.gender_restriction, e.age_min, e.age_max,
		       e.smoking_allowed, e.alcohol_allowed, e.event_languages, e.slug, e.created_at,
		       u.email
		FROM events e
		LEFT JOIN users u ON e.user_id = u.id
		WHERE e.id = ?
	`, id).Scan(
		&e.ID, &e.UserID, &e.Title, &e.Description, &e.Category, &e.Latitude, &e.Longitude,
		&startTime, &endTime, &e.CreatorName,
		&maxParticipants, &genderRestriction, &e.AgeMin, &e.AgeMax,
		&e.SmokingAllowed, &e.AlcoholAllowed, &eventLanguages, &slug, &createdAt, &e.UserEmail,
	)

	if err == sql.ErrNoRows {
		log.Printf("❌ Event %s not found", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
		log.Printf("❌ Error fetching event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	if startTime.Valid {
		e.StartTime = startTime.String
	}
	if endTime.Valid {
		e.EndTime = endTime.String
	}
	if maxParticipants.Valid {
		e.MaxParticipants = int(maxParticipants.Int64)
	}
	if genderRestriction.Valid {
		e.GenderRestriction = genderRestriction.String
	} else {
		e.GenderRestriction = "any"
	}
	if eventLanguages.Valid {
		e.EventLanguages = eventLanguages.String
	}
	if slug.Valid {
		e.Slug = slug.String
	}
	e.CreatedAt = createdAt

	log.Printf("✓ Event %s found", id)
	c.JSON(http.StatusOK, e)
}

func createEvent(c *gin.Context) {
	userID := c.GetInt("user_id")
	requestID, _ := c.Get("request_id")
	log.Printf("[%v] ➕ POST /api/events - Creating new event for user ID: %d", requestID, userID)

	// Check if user's email is verified (admins are exempt)
	var emailVerified, isAdmin bool
	err := db.QueryRow(`SELECT email_verified, is_admin FROM users WHERE id = ?`, userID).Scan(&emailVerified, &isAdmin)
	if err != nil {
		log.Printf("[%v] ❌ Failed to check email verification status: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify account status"})
		return
	}
	// Admins can create events without email verification (they can verify themselves)
	if !emailVerified && !isAdmin {
		log.Printf("[%v] ❌ User %d attempted to create event with unverified email", requestID, userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your email address before creating events"})
		return
	}

	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		log.Printf("[%v] ❌ Invalid JSON: %v", requestID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Parse times first
	startTime, err := parseDateTime(event.StartTime)
	if err != nil {
		log.Printf("[%v] ❌ Invalid start_time: %v", requestID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format"})
		return
	}

	var endTimePtr *time.Time
	if event.EndTime != "" {
		endTime, err := parseDateTime(event.EndTime)
		if err != nil {
			log.Printf("[%v] ❌ Invalid end_time: %v", requestID, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format"})
			return
		}
		endTimePtr = &endTime
	}

	// Validate event data
	if err := ValidateEvent(&event, &startTime, endTimePtr); err != nil {
		log.Printf("[%v] ❌ Validation failed: %v", requestID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate unique slug for the event (with uniqueness check)
	slug, err := generateUniqueSlug(event.Title)
	if err != nil {
		log.Printf("[%v] ❌ Slug generation failed: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate event URL"})
		return
	}
	log.Printf("✓ Generated slug: %s", slug)

	result, err := db.Exec(`
		INSERT INTO events (
			user_id, title, description, category, latitude, longitude, start_time, end_time,
			creator_name, max_participants,
			gender_restriction, age_min, age_max,
			smoking_allowed, alcohol_allowed, event_languages, slug,
			hide_organizer_until_joined, hide_participants_until_joined,
			require_verified_to_join, require_verified_to_view
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, event.Title, event.Description, event.Category, event.Latitude, event.Longitude,
		startTime, endTimePtr, event.CreatorName,
		event.MaxParticipants, event.GenderRestriction, event.AgeMin, event.AgeMax,
		event.SmokingAllowed, event.AlcoholAllowed, event.EventLanguages, slug,
		event.HideOrganizerUntilJoined, event.HideParticipantsUntilJoined,
		event.RequireVerifiedToJoin, event.RequireVerifiedToView)

	if err != nil {
		log.Printf("❌ Database insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("❌ Failed to get event ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}
	event.ID = int(id)
	event.UserID = userID
	event.Slug = slug
	event.CreatedAt = time.Now()

	log.Printf("✅ Event created successfully with ID: %d, slug: %s", id, slug)
	c.JSON(http.StatusCreated, event)
}

func updateEvent(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetInt("user_id")
	isAdmin := c.GetBool("is_admin")

	log.Printf("✏️ PUT /api/events/%s - Updating event", id)

	// Check ownership
	var eventUserID int
	err := db.QueryRow("SELECT user_id FROM events WHERE id = ?", id).Scan(&eventUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if eventUserID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this event"})
		return
	}

	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		log.Printf("❌ Invalid JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	startTime, err := parseDateTime(event.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid start_time: %v", err)})
		return
	}

	var endTimePtr *time.Time
	if event.EndTime != "" {
		endTime, err := parseDateTime(event.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid end_time: %v", err)})
			return
		}
		endTimePtr = &endTime
	}

	result, err := db.Exec(`
		UPDATE events SET
			title = ?, description = ?, category = ?, latitude = ?, longitude = ?,
			start_time = ?, end_time = ?, creator_name = ?,
			max_participants = ?, gender_restriction = ?, age_min = ?, age_max = ?,
			smoking_allowed = ?, alcohol_allowed = ?, event_languages = ?
		WHERE id = ?
	`, event.Title, event.Description, event.Category, event.Latitude, event.Longitude,
		startTime, endTimePtr, event.CreatorName,
		event.MaxParticipants, event.GenderRestriction, event.AgeMin, event.AgeMax,
		event.SmokingAllowed, event.AlcoholAllowed, event.EventLanguages, id)

	if err != nil {
		log.Printf("❌ Database update failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	eventID, _ := strconv.Atoi(id)
	event.ID = eventID
	log.Printf("✅ Event %s updated successfully", id)
	c.JSON(http.StatusOK, event)
}

func deleteEvent(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetInt("user_id")
	isAdmin := c.GetBool("is_admin")

	log.Printf("🗑️ DELETE /api/events/%s - Deleting event", id)

	// Check ownership
	var eventUserID int
	err := db.QueryRow("SELECT user_id FROM events WHERE id = ?", id).Scan(&eventUserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if eventUserID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this event"})
		return
	}

	result, err := db.Exec("DELETE FROM events WHERE id = ?", id)
	if err != nil {
		log.Printf("❌ Database delete failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	log.Printf("✅ Event %s deleted successfully", id)
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

func getCategories(c *gin.Context) {
	log.Println("📚 GET /api/categories - Fetching categories")
	c.JSON(http.StatusOK, gin.H{
		"categories": CategoryNames,
	})
}

func searchPlaces(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	log.Printf("🔍 Searching places for: %s", query)

	// Use Photon API from Komoot for geocoding autocomplete
	url := fmt.Sprintf("https://photon.komoot.io/api/?q=%s&limit=10", url.QueryEscape(query))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search places"})
		return
	}

	req.Header.Set("User-Agent", "Veidly/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to fetch places: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search places"})
		return
	}
	defer resp.Body.Close()

	// Photon returns GeoJSON FeatureCollection
	var photonResponse struct {
		Features []struct {
			Geometry struct {
				Coordinates [2]float64 `json:"coordinates"` // [lon, lat]
			} `json:"geometry"`
			Properties struct {
				Name        string `json:"name"`
				City        string `json:"city"`
				Country     string `json:"country"`
				Street      string `json:"street"`
				Housenumber string `json:"housenumber"`
				State       string `json:"state"`
				Postcode    string `json:"postcode"`
				OSMType     string `json:"osm_type"`
				OSMValue    string `json:"osm_value"`
			} `json:"properties"`
		} `json:"features"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&photonResponse); err != nil {
		log.Printf("❌ Failed to decode response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse places"})
		return
	}

	// Convert Photon response to our Place format
	places := make([]Place, 0, len(photonResponse.Features))
	for _, feature := range photonResponse.Features {
		// Build display name from available properties
		displayName := buildDisplayName(feature.Properties)

		places = append(places, Place{
			DisplayName: displayName,
			Lat:         fmt.Sprintf("%f", feature.Geometry.Coordinates[1]),
			Lon:         fmt.Sprintf("%f", feature.Geometry.Coordinates[0]),
			Type:        feature.Properties.OSMValue,
			Importance:  0, // Photon doesn't provide importance, but order is by relevance
		})
	}

	log.Printf("✓ Found %d places", len(places))
	c.JSON(http.StatusOK, places)
}

// buildDisplayName creates a human-readable display name from Photon properties
func buildDisplayName(props struct {
	Name        string `json:"name"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Street      string `json:"street"`
	Housenumber string `json:"housenumber"`
	State       string `json:"state"`
	Postcode    string `json:"postcode"`
	OSMType     string `json:"osm_type"`
	OSMValue    string `json:"osm_value"`
}) string {
	parts := []string{}

	// Add street address if available
	if props.Housenumber != "" && props.Street != "" {
		parts = append(parts, fmt.Sprintf("%s %s", props.Street, props.Housenumber))
	} else if props.Street != "" {
		parts = append(parts, props.Street)
	} else if props.Name != "" {
		parts = append(parts, props.Name)
	}

	// Add city
	if props.City != "" {
		parts = append(parts, props.City)
	}

	// Add state if available and different from city
	if props.State != "" && props.State != props.City {
		parts = append(parts, props.State)
	}

	// Add country
	if props.Country != "" {
		parts = append(parts, props.Country)
	}

	// If no parts were added, use name or fallback
	if len(parts) == 0 {
		if props.Name != "" {
			return props.Name
		}
		return "Unknown Location"
	}

	return strings.Join(parts, ", ")
}

// Admin handlers
func adminGetUsers(c *gin.Context) {
	log.Println("👥 GET /api/admin/users - Admin fetching all users")

	rows, err := db.Query(`
		SELECT id, email, name, bio, languages, is_admin, is_blocked, email_verified, created_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var bio, languages sql.NullString
		err := rows.Scan(&u.ID, &u.Email, &u.Name, &bio, &languages, &u.IsAdmin, &u.IsBlocked, &u.EmailVerified, &u.CreatedAt)
		if err != nil {
			continue
		}
		// Convert NullString to string
		if bio.Valid {
			u.Bio = bio.String
		}
		if languages.Valid {
			u.Languages = languages.String
		}
		users = append(users, u)
	}

	log.Printf("✓ Found %d users", len(users))
	c.JSON(http.StatusOK, users)
}

func adminBlockUser(c *gin.Context) {
	id := c.Param("id")
	log.Printf("🚫 PUT /api/admin/users/%s/block - Admin blocking user", id)

	_, err := db.Exec("UPDATE users SET is_blocked = 1 WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to block user"})
		return
	}

	log.Printf("✅ User %s blocked", id)
	c.JSON(http.StatusOK, gin.H{"message": "User blocked successfully"})
}

func adminUnblockUser(c *gin.Context) {
	id := c.Param("id")
	log.Printf("✅ PUT /api/admin/users/%s/unblock - Admin unblocking user", id)

	_, err := db.Exec("UPDATE users SET is_blocked = 0 WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock user"})
		return
	}

	log.Printf("✅ User %s unblocked", id)
	c.JSON(http.StatusOK, gin.H{"message": "User unblocked successfully"})
}

func adminVerifyUserEmail(c *gin.Context) {
	id := c.Param("id")
	log.Printf("✉️  PUT /api/admin/users/%s/verify-email - Admin manually verifying user email", id)

	_, err := db.Exec("UPDATE users SET email_verified = 1 WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user email"})
		return
	}

	log.Printf("✅ User %s email verified by admin", id)
	c.JSON(http.StatusOK, gin.H{"message": "User email verified successfully"})
}

func adminGetAllEvents(c *gin.Context) {
	log.Println("📋 GET /api/admin/events - Admin fetching all events")

	rows, err := db.Query(`
		SELECT e.id, e.user_id, e.title, e.description, e.category, e.latitude, e.longitude,
		       e.start_time, e.end_time, e.creator_name, e.max_participants,
		       e.gender_restriction, e.age_min, e.age_max,
		       e.smoking_allowed, e.alcohol_allowed, e.event_languages, e.slug, e.created_at,
		       u.email
		FROM events e
		LEFT JOIN users u ON e.user_id = u.id
		ORDER BY e.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var startTime, endTime, eventLanguages, slug sql.NullString
		var createdAt time.Time
		err := rows.Scan(
			&e.ID, &e.UserID, &e.Title, &e.Description, &e.Category, &e.Latitude, &e.Longitude,
			&startTime, &endTime, &e.CreatorName,
			&e.MaxParticipants, &e.GenderRestriction, &e.AgeMin, &e.AgeMax,
			&e.SmokingAllowed, &e.AlcoholAllowed, &eventLanguages, &slug, &createdAt, &e.UserEmail,
		)
		if err != nil {
			continue
		}
		if startTime.Valid {
			e.StartTime = startTime.String
		}
		if endTime.Valid {
			e.EndTime = endTime.String
		}
		if eventLanguages.Valid {
			e.EventLanguages = eventLanguages.String
		}
		if slug.Valid {
			e.Slug = slug.String
		}
		e.CreatedAt = createdAt
		events = append(events, e)
	}

	log.Printf("✓ Found %d events", len(events))
	c.JSON(http.StatusOK, events)
}

func adminDeleteEvent(c *gin.Context) {
	id := c.Param("id")
	log.Printf("🗑️ DELETE /api/admin/events/%s - Admin deleting event", id)

	result, err := db.Exec("DELETE FROM events WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	log.Printf("✅ Event %s deleted by admin", id)
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

func adminUpdateEvent(c *gin.Context) {
	id := c.Param("id")
	log.Printf("✏️ PUT /api/admin/events/%s - Admin updating event", id)

	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	startTime, err := parseDateTime(event.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid start_time: %v", err)})
		return
	}

	var endTimePtr *time.Time
	if event.EndTime != "" {
		endTime, err := parseDateTime(event.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid end_time: %v", err)})
			return
		}
		endTimePtr = &endTime
	}

	result, err := db.Exec(`
		UPDATE events SET
			title = ?, description = ?, category = ?, latitude = ?, longitude = ?,
			start_time = ?, end_time = ?, creator_name = ?,
			max_participants = ?, gender_restriction = ?, age_min = ?, age_max = ?,
			smoking_allowed = ?, alcohol_allowed = ?, event_languages = ?
		WHERE id = ?
	`, event.Title, event.Description, event.Category, event.Latitude, event.Longitude,
		startTime, endTimePtr, event.CreatorName,
		event.MaxParticipants, event.GenderRestriction, event.AgeMin, event.AgeMax,
		event.SmokingAllowed, event.AlcoholAllowed, event.EventLanguages, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	log.Printf("✅ Event %s updated by admin", id)
	c.JSON(http.StatusOK, event)
}

// Profile handlers
func getOwnProfile(c *gin.Context) {
	userID := c.GetInt("user_id")
	log.Printf("👤 GET /api/profile - Fetching own profile for user ID: %d", userID)

	// Get user profile
	var user User
	var bio, languages sql.NullString
	err := db.QueryRow(`
		SELECT id, email, name, bio, languages, is_admin, is_blocked, email_verified, created_at
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Email, &user.Name, &bio, &languages,
		&user.IsAdmin, &user.IsBlocked, &user.EmailVerified, &user.CreatedAt)

	// Convert NullString to string
	if bio.Valid {
		user.Bio = bio.String
	}
	if languages.Valid {
		user.Languages = languages.String
	}

	if err != nil {
		log.Printf("❌ Failed to fetch user profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
		return
	}

	// Get user's created events (upcoming)
	createdRows, err := db.Query(`
		SELECT id, title, slug, start_time, category, latitude, longitude
		FROM events
		WHERE user_id = ? AND start_time > datetime('now')
		ORDER BY start_time ASC
	`, userID)

	if err != nil {
		log.Printf("❌ Failed to fetch created events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}
	defer createdRows.Close()

	createdEvents := []map[string]interface{}{}
	for createdRows.Next() {
		var id int
		var title, slug, startTime, category string
		var lat, lng float64
		if err := createdRows.Scan(&id, &title, &slug, &startTime, &category, &lat, &lng); err != nil {
			log.Printf("❌ Error scanning created event: %v", err)
			continue
		}
		createdEvents = append(createdEvents, map[string]interface{}{
			"id":         id,
			"title":      title,
			"slug":       slug,
			"start_time": startTime,
			"category":   category,
			"latitude":   lat,
			"longitude":  lng,
		})
	}

	// Get user's joined events (upcoming, not created by user)
	joinedRows, err := db.Query(`
		SELECT e.id, e.title, e.slug, e.start_time, e.category, e.latitude, e.longitude
		FROM events e
		INNER JOIN event_participants ep ON e.id = ep.event_id
		WHERE ep.user_id = ? AND e.user_id != ? AND e.start_time > datetime('now')
		ORDER BY e.start_time ASC
	`, userID, userID)

	if err != nil {
		log.Printf("❌ Failed to fetch joined events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch joined events"})
		return
	}
	defer joinedRows.Close()

	joinedEvents := []map[string]interface{}{}
	for joinedRows.Next() {
		var id int
		var title, slug, startTime, category string
		var lat, lng float64
		if err := joinedRows.Scan(&id, &title, &slug, &startTime, &category, &lat, &lng); err != nil {
			log.Printf("❌ Error scanning joined event: %v", err)
			continue
		}
		joinedEvents = append(joinedEvents, map[string]interface{}{
			"id":         id,
			"title":      title,
			"slug":       slug,
			"start_time": startTime,
			"category":   category,
			"latitude":   lat,
			"longitude":  lng,
		})
	}

	// Get user's past events (both created and joined)
	pastRows, err := db.Query(`
		SELECT DISTINCT e.id, e.title, e.slug, e.start_time, e.category, e.latitude, e.longitude,
		CASE WHEN e.user_id = ? THEN 1 ELSE 0 END as is_creator
		FROM events e
		LEFT JOIN event_participants ep ON e.id = ep.event_id
		WHERE (e.user_id = ? OR ep.user_id = ?) AND e.start_time <= datetime('now')
		ORDER BY e.start_time DESC
	`, userID, userID, userID)

	if err != nil {
		log.Printf("❌ Failed to fetch past events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch past events"})
		return
	}
	defer pastRows.Close()

	pastEvents := []map[string]interface{}{}
	for pastRows.Next() {
		var id int
		var title, slug, startTime, category string
		var lat, lng float64
		var isCreator int
		if err := pastRows.Scan(&id, &title, &slug, &startTime, &category, &lat, &lng, &isCreator); err != nil {
			log.Printf("❌ Error scanning past event: %v", err)
			continue
		}
		pastEvents = append(pastEvents, map[string]interface{}{
			"id":         id,
			"title":      title,
			"slug":       slug,
			"start_time": startTime,
			"category":   category,
			"latitude":   lat,
			"longitude":  lng,
			"is_creator": isCreator == 1,
		})
	}

	log.Printf("✅ Profile fetched for user: %s with %d created, %d joined, %d past events",
		user.Email, len(createdEvents), len(joinedEvents), len(pastEvents))
	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"created_events": createdEvents,
		"joined_events":  joinedEvents,
		"past_events":    pastEvents,
	})
}

func updateProfile(c *gin.Context) {
	userID := c.GetInt("user_id")
	log.Printf("👤 PUT /api/profile - Updating profile for user ID: %d", userID)

	var req ProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	_, err := db.Exec(`
		UPDATE users SET name = ?, bio = ?, languages = ?
		WHERE id = ?
	`, req.Name, req.Bio, req.Languages, userID)

	if err != nil {
		log.Printf("❌ Profile update failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Get updated user
	var user User
	var bio, languages sql.NullString
	err = db.QueryRow(`
		SELECT id, email, name, bio, languages, is_admin, is_blocked, email_verified, created_at
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Email, &user.Name, &bio, &languages,
		&user.IsAdmin, &user.IsBlocked, &user.EmailVerified, &user.CreatedAt)

	// Convert NullString to string
	if bio.Valid {
		user.Bio = bio.String
	}
	if languages.Valid {
		user.Languages = languages.String
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated profile"})
		return
	}

	log.Printf("✅ Profile updated for user: %s", user.Email)
	c.JSON(http.StatusOK, user)
}

func getUserProfile(c *gin.Context) {
	id := c.Param("id")
	log.Printf("👤 GET /api/profile/%s - Fetching user profile", id)

	var user User
	var bio, languages sql.NullString
	err := db.QueryRow(`
		SELECT id, email, name, bio, languages, is_admin, is_blocked, email_verified, created_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Email, &user.Name, &bio, &languages,
		&user.IsAdmin, &user.IsBlocked, &user.EmailVerified, &user.CreatedAt)

	// Convert NullString to string
	if bio.Valid {
		user.Bio = bio.String
	}
	if languages.Valid {
		user.Languages = languages.String
	}

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user profile"})
		return
	}

	userIDInt, _ := strconv.Atoi(id)

	// Get user's created events (upcoming only for other users)
	createdRows, err := db.Query(`
		SELECT id, title, slug, start_time, category, latitude, longitude
		FROM events
		WHERE user_id = ? AND start_time > datetime('now')
		ORDER BY start_time ASC
	`, userIDInt)

	if err != nil {
		log.Printf("❌ Failed to fetch created events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}
	defer createdRows.Close()

	createdEvents := []map[string]interface{}{}
	for createdRows.Next() {
		var eventID int
		var title, slug, startTime, category string
		var lat, lng float64
		if err := createdRows.Scan(&eventID, &title, &slug, &startTime, &category, &lat, &lng); err != nil {
			log.Printf("❌ Error scanning created event: %v", err)
			continue
		}
		createdEvents = append(createdEvents, map[string]interface{}{
			"id":         eventID,
			"title":      title,
			"slug":       slug,
			"start_time": startTime,
			"category":   category,
			"latitude":   lat,
			"longitude":  lng,
		})
	}

	log.Printf("✓ Profile found for: %s with %d upcoming events", user.Email, len(createdEvents))
	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"created_events": createdEvents,
	})
}

// Event participation handlers
func joinEvent(c *gin.Context) {
	eventID := c.Param("id")
	userID := c.GetInt("user_id")
	isVerified := c.GetBool("email_verified")
	isAdmin := c.GetBool("is_admin")

	log.Printf("➕ POST /api/events/%s/join - User %d joining event", eventID, userID)

	// GLOBAL REQUIREMENT: Email verification required for all event joins (except admins)
	if !isVerified && !isAdmin {
		log.Printf("❌ User %d needs verified email to join any event", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "You must verify your email address before joining events. Please check your email for the verification link."})
		return
	}

	// Start transaction to prevent race condition (CRITICAL SECURITY FIX)
	// Without transaction, multiple users could join simultaneously when only 1 spot left
	tx, err := db.Begin()
	if err != nil {
		log.Printf("❌ Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join event"})
		return
	}
	defer tx.Rollback() // Will be no-op if tx.Commit() succeeds

	// Check if event exists, has space, and check privacy settings WITH ROW LOCK
	var maxParticipants sql.NullInt64
	var currentCount int
	var requireVerifiedToJoin bool
	err = tx.QueryRow(`
		SELECT max_participants,
		       (SELECT COUNT(*) FROM event_participants WHERE event_id = ?) as count,
		       require_verified_to_join
		FROM events WHERE id = ?
	`, eventID, eventID).Scan(&maxParticipants, &currentCount, &requireVerifiedToJoin)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
		log.Printf("❌ Error checking event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join event"})
		return
	}

	// The require_verified_to_join flag is now redundant (kept for backward compatibility)
	// but the global check above already enforces verification for all events

	// Check capacity
	if maxParticipants.Valid && currentCount >= int(maxParticipants.Int64) {
		log.Printf("❌ Event %s is full (%d/%d participants)", eventID, currentCount, maxParticipants.Int64)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event is full"})
		return
	}

	// Insert participant within transaction
	_, err = tx.Exec(`
		INSERT INTO event_participants (event_id, user_id)
		VALUES (?, ?)
	`, eventID, userID)

	// Handle duplicate join (UNIQUE constraint)
	if err != nil {
		log.Printf("❌ User %d already joined event %s", userID, eventID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already joined this event"})
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("❌ Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join event"})
		return
	}

	log.Printf("✅ User %d successfully joined event %s", userID, eventID)
	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined event"})
}

func leaveEvent(c *gin.Context) {
	eventID := c.Param("id")
	userID := c.GetInt("user_id")

	log.Printf("➖ DELETE /api/events/%s/leave - User %d leaving event", eventID, userID)

	result, err := db.Exec(`
		DELETE FROM event_participants
		WHERE event_id = ? AND user_id = ?
	`, eventID, userID)

	if err != nil {
		log.Printf("❌ Error leaving event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave event"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		log.Printf("❌ User %d is not a participant of event %s", userID, eventID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a participant of this event"})
		return
	}

	log.Printf("✅ User %d successfully left event %s", userID, eventID)
	c.JSON(http.StatusOK, gin.H{"message": "Successfully left event"})
}

func getPublicEvent(c *gin.Context) {
	slug := c.Param("slug")
	log.Printf("🌐 GET /api/public/events/%s - Fetching public event by slug", slug)

	// Get viewer info for privacy filtering
	viewerUserID, _ := c.Get("user_id")
	viewerIsAdmin, _ := c.Get("is_admin")
	viewerIsVerified, _ := c.Get("email_verified")

	userID := 0
	isAdmin := false
	isVerified := false

	if viewerUserID != nil {
		userID, _ = viewerUserID.(int)
	}
	if viewerIsAdmin != nil {
		isAdmin, _ = viewerIsAdmin.(bool)
	}
	if viewerIsVerified != nil {
		isVerified, _ = viewerIsVerified.(bool)
	}

	var e Event
	var startTime, endTime, genderRestriction, eventLanguages, eventSlug, userEmail, creatorLanguages sql.NullString
	var maxParticipants sql.NullInt64
	var createdAt time.Time
	var isParticipant bool

	// Build query with participant check if user is authenticated
	query := `
		SELECT e.id, e.user_id, e.title, e.description, e.category, e.latitude, e.longitude,
		       e.start_time, e.end_time, e.creator_name, e.max_participants,
		       e.gender_restriction, e.age_min, e.age_max,
		       e.smoking_allowed, e.alcohol_allowed, e.event_languages, e.slug, e.created_at,
		       e.hide_organizer_until_joined, e.hide_participants_until_joined,
		       e.require_verified_to_join, e.require_verified_to_view,
		       u.email, u.languages as creator_languages,
		       (SELECT COUNT(*) FROM event_participants WHERE event_id = e.id) as participant_count
	`

	var err error
	if userID > 0 {
		query += `, (SELECT COUNT(*) > 0 FROM event_participants WHERE event_id = e.id AND user_id = ?) as is_participant
			FROM events e
			LEFT JOIN users u ON e.user_id = u.id
			WHERE e.slug = ?`
		err = db.QueryRow(query, userID, slug).Scan(
			&e.ID, &e.UserID, &e.Title, &e.Description, &e.Category, &e.Latitude, &e.Longitude,
			&startTime, &endTime, &e.CreatorName,
			&maxParticipants, &genderRestriction, &e.AgeMin, &e.AgeMax,
			&e.SmokingAllowed, &e.AlcoholAllowed, &eventLanguages, &eventSlug, &createdAt,
			&e.HideOrganizerUntilJoined, &e.HideParticipantsUntilJoined,
			&e.RequireVerifiedToJoin, &e.RequireVerifiedToView,
			&userEmail, &creatorLanguages, &e.ParticipantCount, &isParticipant,
		)
	} else {
		query += `, 0 as is_participant
			FROM events e
			LEFT JOIN users u ON e.user_id = u.id
			WHERE e.slug = ?`
		err = db.QueryRow(query, slug).Scan(
			&e.ID, &e.UserID, &e.Title, &e.Description, &e.Category, &e.Latitude, &e.Longitude,
			&startTime, &endTime, &e.CreatorName,
			&maxParticipants, &genderRestriction, &e.AgeMin, &e.AgeMax,
			&e.SmokingAllowed, &e.AlcoholAllowed, &eventLanguages, &eventSlug, &createdAt,
			&e.HideOrganizerUntilJoined, &e.HideParticipantsUntilJoined,
			&e.RequireVerifiedToJoin, &e.RequireVerifiedToView,
			&userEmail, &creatorLanguages, &e.ParticipantCount, &isParticipant,
		)
	}

	if err == sql.ErrNoRows {
		log.Printf("❌ Event with slug %s not found", slug)
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
		log.Printf("❌ Error fetching event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	if startTime.Valid {
		e.StartTime = startTime.String
	}
	if endTime.Valid {
		e.EndTime = endTime.String
	}
	if maxParticipants.Valid {
		e.MaxParticipants = int(maxParticipants.Int64)
	}
	if genderRestriction.Valid {
		e.GenderRestriction = genderRestriction.String
	} else {
		e.GenderRestriction = "any"
	}
	if eventLanguages.Valid {
		e.EventLanguages = eventLanguages.String
	}
	if eventSlug.Valid {
		e.Slug = eventSlug.String
	}
	if userEmail.Valid {
		e.UserEmail = userEmail.String
	}
	if creatorLanguages.Valid {
		e.CreatorLanguages = creatorLanguages.String
	}
	e.CreatedAt = createdAt
	e.IsParticipant = isParticipant

	// Check if event can be viewed
	if errMsg := CheckEventViewPermission(&e, userID, isVerified, isAdmin); errMsg != "" {
		log.Printf("❌ User cannot view event %s: %s", slug, errMsg)
		c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		return
	}

	// Apply privacy filters
	ApplyPrivacyFilters(&e, userID, isVerified, isAdmin)

	log.Printf("✓ Public event found: %s (ID: %d)", slug, e.ID)
	c.JSON(http.StatusOK, e)
}

func getEventParticipants(c *gin.Context) {
	eventID := c.Param("id")
	eventIDInt, err := strconv.Atoi(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	log.Printf("👥 GET /api/events/%s/participants - Fetching participants", eventID)

	// Get viewer info for privacy filtering
	viewerUserID, _ := c.Get("user_id")
	viewerIsAdmin, _ := c.Get("is_admin")
	viewerIsVerified, _ := c.Get("email_verified")

	userID := 0
	isAdmin := false
	isVerified := false

	if viewerUserID != nil {
		userID, _ = viewerUserID.(int)
	}
	if viewerIsAdmin != nil {
		isAdmin, _ = viewerIsAdmin.(bool)
	}
	if viewerIsVerified != nil {
		isVerified, _ = viewerIsVerified.(bool)
	}

	// Use privacy-aware participant fetching
	participants, err := GetParticipantsWithPrivacy(eventIDInt, userID, isVerified, isAdmin)
	if err != nil {
		log.Printf("❌ Error fetching participants: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve participants"})
		return
	}

	log.Printf("✓ Found %d participants for event %s", len(participants), eventID)
	c.JSON(http.StatusOK, participants)
}

func downloadEventICS(c *gin.Context) {
	slug := c.Param("slug")
	log.Printf("📅 GET /api/public/events/%s/ics - Downloading ICS file", slug)

	var e Event
	var startTime, endTime, genderRestriction, eventLanguages, eventSlug sql.NullString
	var maxParticipants sql.NullInt64
	var createdAt time.Time
	
	err := db.QueryRow(`
		SELECT e.id, e.user_id, e.title, e.description, e.category, e.latitude, e.longitude,
		       e.start_time, e.end_time, e.creator_name, e.max_participants,
		       e.gender_restriction, e.age_min, e.age_max,
		       e.smoking_allowed, e.alcohol_allowed, e.event_languages, e.slug, e.created_at
		FROM events e
		WHERE e.slug = ?
	`, slug).Scan(
		&e.ID, &e.UserID, &e.Title, &e.Description, &e.Category, &e.Latitude, &e.Longitude,
		&startTime, &endTime, &e.CreatorName,
		&maxParticipants, &genderRestriction, &e.AgeMin, &e.AgeMax,
		&e.SmokingAllowed, &e.AlcoholAllowed, &eventLanguages, &eventSlug, &createdAt,
	)

	if err == sql.ErrNoRows {
		log.Printf("❌ Event with slug %s not found", slug)
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
		log.Printf("❌ Error fetching event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	if startTime.Valid {
		e.StartTime = startTime.String
	}
	if endTime.Valid {
		e.EndTime = endTime.String
	}
	if maxParticipants.Valid {
		e.MaxParticipants = int(maxParticipants.Int64)
	}
	if genderRestriction.Valid {
		e.GenderRestriction = genderRestriction.String
	}
	if eventLanguages.Valid {
		e.EventLanguages = eventLanguages.String
	}
	if eventSlug.Valid {
		e.Slug = eventSlug.String
	}
	e.CreatedAt = createdAt

	// Generate ICS content
	icsContent := GenerateICS(&e)

	// Set headers for file download
	filename := fmt.Sprintf("%s.ics", slug)
	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Cache-Control", "no-cache")

	log.Printf("✅ ICS file generated for event: %s", slug)
	c.String(http.StatusOK, icsContent)
}
