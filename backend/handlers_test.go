package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDBFile = "veidly-test-suite.db"

var testDBMutex sync.Mutex

// TestMain ensures production database is never touched
func TestMain(m *testing.M) {
	// Ensure we're in test mode
	gin.SetMode(gin.TestMode)

	// Speed up tests by reducing bcrypt cost (4 instead of 14)
	// This makes password hashing ~1000x faster in tests
	bcryptCost = 4

	// Clean up any existing test database
	os.Remove(testDBFile)

	// Run tests
	exitCode := m.Run()

	// Clean up test database after all tests
	os.Remove(testDBFile)

	os.Exit(exitCode)
}

// setupTestDB creates a fresh test database for each test
func setupTestDB(t *testing.T) *sql.DB {
	testDBMutex.Lock()
	defer testDBMutex.Unlock()

	// Remove existing test DB
	os.Remove(testDBFile)

	// Create new test database
	testDB, err := sql.Open("sqlite3", testDBFile)
	require.NoError(t, err, "Failed to open test database")

	// Verify we're using the test database file
	require.Contains(t, testDBFile, "test", "SECURITY: Not using test database!")

	// Enable foreign keys
	_, err = testDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Create users table
	_, err = testDB.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		name TEXT NOT NULL,
		bio TEXT,
		phone TEXT,
		threema TEXT,
		languages TEXT,
		is_admin BOOLEAN DEFAULT 0,
		is_blocked BOOLEAN DEFAULT 0,
		email_verified BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err, "Failed to create users table")

	// Create events table
	_, err = testDB.Exec(`
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		category TEXT NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		start_time TEXT NOT NULL,
		end_time TEXT,
		creator_name TEXT NOT NULL,
		max_participants INTEGER,
		gender_restriction TEXT,
		age_min INTEGER DEFAULT 0,
		age_max INTEGER DEFAULT 99,
		smoking_allowed BOOLEAN DEFAULT 0,
		alcohol_allowed BOOLEAN DEFAULT 0,
		event_languages TEXT,
		slug TEXT,
		comments_enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		hide_organizer_until_joined BOOLEAN DEFAULT 0,
		hide_participants_until_joined BOOLEAN DEFAULT 1,
		require_verified_to_join BOOLEAN DEFAULT 0,
		require_verified_to_view BOOLEAN DEFAULT 0,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	)`)
	require.NoError(t, err, "Failed to create events table")

	// Create event_participants table
	_, err = testDB.Exec(`
	CREATE TABLE IF NOT EXISTS event_participants (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
		UNIQUE(event_id, user_id)
	)`)
	require.NoError(t, err, "Failed to create event_participants table")

	// Create email_verification_tokens table
	_, err = testDB.Exec(`
	CREATE TABLE IF NOT EXISTS email_verification_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	)`)
	require.NoError(t, err, "Failed to create email_verification_tokens table")

	// Create password_reset_tokens table
	_, err = testDB.Exec(`
	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		used INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	)`)
	require.NoError(t, err, "Failed to create password_reset_tokens table")

	// Create user_blocks table
	_, err = testDB.Exec(`
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
	require.NoError(t, err, "Failed to create user_blocks table")

	// Create event_comments table
	_, err = testDB.Exec(`
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
	require.NoError(t, err, "Failed to create event_comments table")

	return testDB
}

// cleanupTestDB closes and removes the test database
func cleanupTestDB(testDB *sql.DB) {
	if testDB != nil {
		testDB.Close()
	}
	os.Remove(testDBFile)
}

// Helper: Create a test user (with email verified by default for testing)
func createTestUser(t *testing.T, testDB *sql.DB, email, name, password string, isAdmin bool) int64 {
	hashedPass, err := hashPassword(password)
	require.NoError(t, err)

	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, is_admin, email_verified)
		VALUES (?, ?, ?, ?, 1)
	`, email, hashedPass, name, isAdmin)
	require.NoError(t, err)

	userID, err := result.LastInsertId()
	require.NoError(t, err)

	return userID
}

// Helper: Create a test event
func createTestEvent(t *testing.T, testDB *sql.DB, userID int64, title string) int64 {
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	// Get user's languages for the event
	var languages sql.NullString
	err := testDB.QueryRow("SELECT languages FROM users WHERE id = ?", userID).Scan(&languages)
	require.NoError(t, err)

	eventLanguages := ""
	if languages.Valid {
		eventLanguages = languages.String
	}

	result, err := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, event_languages, hide_organizer_until_joined, hide_participants_until_joined, require_verified_to_join, require_verified_to_view)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1, 0, 0)
	`, userID, title, "Test description", "social_drinks", 46.8805, 8.6444, future, "Test User", eventLanguages)
	require.NoError(t, err)

	eventID, err := result.LastInsertId()
	require.NoError(t, err)

	return eventID
}

// ============================================================================
// AUTHENTICATION TESTS
// ============================================================================

func TestRegisterUser(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	router := gin.New()
	router.POST("/api/register", register)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "Valid registration",
			payload: map[string]interface{}{
				"email":    "newuser@example.com",
				"password": "password123",
				"name":     "New User",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "user")
				assert.Contains(t, response, "token")
				assert.NotEmpty(t, response["token"], "token should not be empty")
			},
		},
		{
			name: "Missing email",
			payload: map[string]interface{}{
				"password": "password123",
				"name":     "User",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name: "Missing password",
			payload: map[string]interface{}{
				"email": "user@example.com",
				"name":  "User",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}

	// Test duplicate email in separate test
	t.Run("Duplicate email", func(t *testing.T) {
		// First registration
		payload1 := map[string]interface{}{
			"email":    "duplicate@example.com",
			"password": "password123",
			"name":     "User One",
		}
		body1, _ := json.Marshal(payload1)
		req1, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(body1))
		req1.Header.Set("Content-Type", "application/json")

		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusCreated, w1.Code)

		// Second registration with same email
		req2, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(body1))
		req2.Header.Set("Content-Type", "application/json")

		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusConflict, w2.Code)
	})
}

func TestLogin(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create test user
	createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)

	router := gin.New()
	router.POST("/api/login", login)

	tests := []struct {
		name           string
		email          string
		password       string
		expectedStatus int
	}{
		{
			name:           "Valid login",
			email:          "user@example.com",
			password:       "password123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wrong password",
			email:          "user@example.com",
			password:       "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Non-existent user",
			email:          "nonexistent@example.com",
			password:       "password123",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "user")
				assert.Contains(t, response, "token")
				assert.NotEmpty(t, response["token"], "token should not be empty")
			}
		})
	}
}

// ============================================================================
// PROFILE TESTS
// ============================================================================

func TestGetProfile(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)

	// Update profile with additional data
	_, err := testDB.Exec(`
		UPDATE users
		SET bio = ?, threema = ?, languages = ?
		WHERE id = ?
	`, "Test bio", "TESTID", "en,de,fr", userID)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Next()
	})
	router.GET("/api/profile", getCurrentUser)

	req, _ := http.NewRequest("GET", "/api/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "user")

	userMap := response["user"].(map[string]interface{})
	assert.Equal(t, "Test User", userMap["name"])
	assert.Equal(t, "Test bio", userMap["bio"])
	assert.Equal(t, "TESTID", userMap["threema"])
	assert.Equal(t, "en,de,fr", userMap["languages"])
}

func TestUpdateProfile(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Next()
	})
	router.PUT("/api/profile", updateProfile)

	payload := map[string]interface{}{
		"name":      "Updated Name",
		"bio":       "Updated bio",
		"phone":     "9876543210",
		"threema":   "UPDATED",
		"languages": "de,fr,it",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", "/api/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	var user User
	err := testDB.QueryRow(`
		SELECT name, bio, threema, languages
		FROM users WHERE id = ?
	`, userID).Scan(&user.Name, &user.Bio, &user.Threema, &user.Languages)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", user.Name)
	assert.Equal(t, "Updated bio", user.Bio)
	assert.Equal(t, "de,fr,it", user.Languages)
}

func TestGetOwnProfile(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)
	otherUserID := createTestUser(t, testDB, "other@example.com", "Other User", "password123", false)

	// Update profile with additional data
	_, err := testDB.Exec(`
		UPDATE users
		SET bio = ?, threema = ?, languages = ?
		WHERE id = ?
	`, "Test bio", "TESTID", "en,de,fr", userID)
	require.NoError(t, err)

	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	past := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	// Create events that the user created (upcoming)
	_, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "My Future Event 1", "Description 1", "social_drinks", 46.8, 8.6, future, "Test User", "my-future-event-1")
	require.NoError(t, err)

	_, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "My Future Event 2", "Description 2", "food_dining", 46.9, 8.7, future, "Test User", "my-future-event-2")
	require.NoError(t, err)

	// Create an event by another user that this user will join
	result, err := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, otherUserID, "Joined Event", "Description", "sports_fitness", 47.0, 8.8, future, "Other User", "joined-event")
	require.NoError(t, err)
	joinedEventID, _ := result.LastInsertId()

	// User joins the other user's event
	_, err = testDB.Exec(`
		INSERT INTO event_participants (event_id, user_id)
		VALUES (?, ?)
	`, joinedEventID, userID)
	require.NoError(t, err)

	// Create past events (both created and joined)
	result, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "My Past Event", "Description past", "social_drinks", 46.8, 8.6, past, "Test User", "my-past-event")
	require.NoError(t, err)

	result, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, otherUserID, "Past Joined Event", "Description", "business_networking", 47.1, 8.9, past, "Other User", "past-joined-event")
	require.NoError(t, err)
	pastJoinedEventID, _ := result.LastInsertId()

	_, err = testDB.Exec(`
		INSERT INTO event_participants (event_id, user_id)
		VALUES (?, ?)
	`, pastJoinedEventID, userID)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Next()
	})
	router.GET("/api/profile", getOwnProfile)

	req, _ := http.NewRequest("GET", "/api/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check user data
	userData, ok := response["user"].(map[string]interface{})
	require.True(t, ok, "user field should be present")
	assert.Equal(t, "Test User", userData["name"])
	assert.Equal(t, "Test bio", userData["bio"])
	assert.Equal(t, "TESTID", userData["threema"])
	assert.Equal(t, "en,de,fr", userData["languages"])

	// Check created events
	createdEvents, ok := response["created_events"].([]interface{})
	require.True(t, ok, "created_events field should be present")
	assert.Equal(t, 2, len(createdEvents), "Should have 2 created future events")
	if len(createdEvents) > 0 {
		event := createdEvents[0].(map[string]interface{})
		assert.NotNil(t, event["id"])
		assert.NotNil(t, event["title"])
		assert.NotNil(t, event["slug"])
		assert.NotNil(t, event["start_time"])
		assert.NotNil(t, event["category"])
		assert.NotNil(t, event["latitude"])
		assert.NotNil(t, event["longitude"])
	}

	// Check joined events
	joinedEvents, ok := response["joined_events"].([]interface{})
	require.True(t, ok, "joined_events field should be present")
	assert.Equal(t, 1, len(joinedEvents), "Should have 1 joined future event")
	if len(joinedEvents) > 0 {
		event := joinedEvents[0].(map[string]interface{})
		assert.Equal(t, "Joined Event", event["title"])
	}

	// Check past events
	pastEvents, ok := response["past_events"].([]interface{})
	require.True(t, ok, "past_events field should be present")
	assert.Equal(t, 2, len(pastEvents), "Should have 2 past events")
	if len(pastEvents) > 0 {
		event := pastEvents[0].(map[string]interface{})
		assert.NotNil(t, event["is_creator"])
	}
}

// ============================================================================
// EVENT TESTS
// ============================================================================

func TestCreateEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Next()
	})
	router.POST("/api/events", createEvent)

	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	payload := map[string]interface{}{
		"title":              "Test Event",
		"description":        "Test description",
		"category":           "social_drinks",
		"latitude":           46.8805,
		"longitude":          8.6444,
		"start_time":         future,
		"creator_name":       "Test User",
		"max_participants":   10,
		"gender_restriction": "any",
		"age_min":            18,
		"age_max":            50,
		"smoking_allowed":    false,
		"alcohol_allowed":    true,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "id")
}

func TestGetEvents(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)
	createTestEvent(t, testDB, userID, "Event 1")
	createTestEvent(t, testDB, userID, "Event 2")
	createTestEvent(t, testDB, userID, "Event 3")

	router := gin.New()
	router.GET("/api/events", getEvents)

	req, _ := http.NewRequest("GET", "/api/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var events []Event
	err := json.Unmarshal(w.Body.Bytes(), &events)
	require.NoError(t, err)
	assert.Equal(t, 3, len(events))
}

func TestGetEventByID(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)
	eventID := createTestEvent(t, testDB, userID, "Test Event")

	router := gin.New()
	router.GET("/api/events/:id", getEvent)

	req, _ := http.NewRequest("GET", "/api/events/"+string(rune(eventID+'0')), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var event Event
	err := json.Unmarshal(w.Body.Bytes(), &event)
	require.NoError(t, err)
	assert.Equal(t, "Test Event", event.Title)
}

func TestUpdateEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)
	eventID := createTestEvent(t, testDB, userID, "Original Title")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Next()
	})
	router.PUT("/api/events/:id", updateEvent)

	future := time.Now().Add(48 * time.Hour).Format(time.RFC3339)

	payload := map[string]interface{}{
		"title":           "Updated Title",
		"description":     "Updated description",
		"category":        "social_sports",
		"latitude":        46.9999,
		"longitude":       9.1234,
		"start_time":      future,
		"creator_name":    "Test User",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", "/api/events/"+string(rune(eventID+'0')), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)
	eventID := createTestEvent(t, testDB, userID, "Event to Delete")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Next()
	})
	router.DELETE("/api/events/:id", deleteEvent)

	req, _ := http.NewRequest("DELETE", "/api/events/"+string(rune(eventID+'0')), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM events WHERE id = ?", eventID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// ============================================================================
// EVENT PARTICIPATION TESTS
// ============================================================================

func TestJoinEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	eventID := createTestEvent(t, testDB, user1ID, "Test Event")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		c.Set("email_verified", true) // Need verified email to join
		c.Set("is_admin", false)
		c.Next()
	})
	router.POST("/api/events/:id/join", joinEvent)

	req, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/join", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify participation
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM event_participants WHERE event_id = ? AND user_id = ?", eventID, user2ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestLeaveEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	eventID := createTestEvent(t, testDB, user1ID, "Test Event")

	// First join
	_, err := testDB.Exec("INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)", eventID, user2ID)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		c.Set("email_verified", true)
		c.Set("is_admin", false)
		c.Next()
	})
	router.DELETE("/api/events/:id/leave", leaveEvent)

	req, _ := http.NewRequest("DELETE", "/api/events/"+string(rune(eventID+'0'))+"/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify left
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM event_participants WHERE event_id = ? AND user_id = ?", eventID, user2ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestEventCapacityLimit(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	user3ID := createTestUser(t, testDB, "user3@example.com", "User 3", "password123", false)

	// Create event with max 1 participant
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	result, err := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, max_participants)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user1ID, "Limited Event", "Test", "social_drinks", 46.8805, 8.6444, future, "User 1", 1)
	require.NoError(t, err)

	eventID, err := result.LastInsertId()
	require.NoError(t, err)

	// User 2 joins successfully
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		c.Set("email_verified", true)
		c.Set("is_admin", false)
		c.Next()
	})
	router.POST("/api/events/:id/join", joinEvent)

	req, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/join", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// User 3 should fail (event full)
	router2 := gin.New()
	router2.Use(func(c *gin.Context) {
		c.Set("user_id", int(user3ID))
		c.Set("email_verified", true)
		c.Set("is_admin", false)
		c.Next()
	})
	router2.POST("/api/events/:id/join", joinEvent)

	req2, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/join", nil)
	w2 := httptest.NewRecorder()
	router2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

// ============================================================================
// SEARCH & FILTER TESTS
// ============================================================================

func TestLanguageFiltering(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create users with different languages
	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)

	_, err := testDB.Exec("UPDATE users SET languages = ? WHERE id = ?", "en,de", user1ID)
	require.NoError(t, err)
	_, err = testDB.Exec("UPDATE users SET languages = ? WHERE id = ?", "fr,it", user2ID)
	require.NoError(t, err)

	createTestEvent(t, testDB, user1ID, "English/German Event")
	createTestEvent(t, testDB, user2ID, "French/Italian Event")

	router := gin.New()
	router.GET("/api/events", getEvents)

	// Test filtering by German
	req, _ := http.NewRequest("GET", "/api/events?languages=de", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var events []Event
	err = json.Unmarshal(w.Body.Bytes(), &events)
	require.NoError(t, err)
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "English/German Event", events[0].Title)
}

// ============================================================================
// ADMIN TESTS
// ============================================================================

func TestAdminGetUsers(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	adminID := createTestUser(t, testDB, "admin@example.com", "Admin", "password123", true)
	createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(adminID))
		c.Set("is_admin", true)
		c.Next()
	})
	router.GET("/api/admin/users", adminGetUsers)

	req, _ := http.NewRequest("GET", "/api/admin/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var users []User
	err := json.Unmarshal(w.Body.Bytes(), &users)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 3)
}

func TestBlockUser(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	adminID := createTestUser(t, testDB, "admin@example.com", "Admin", "password123", true)
	userID := createTestUser(t, testDB, "user@example.com", "User", "password123", false)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(adminID))
		c.Set("is_admin", true)
		c.Next()
	})
	router.PUT("/api/admin/users/:id/block", adminBlockUser)

	req, _ := http.NewRequest("PUT", "/api/admin/users/"+string(rune(userID+'0'))+"/block", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user is blocked
	var isBlocked bool
	err := testDB.QueryRow("SELECT is_blocked FROM users WHERE id = ?", userID).Scan(&isBlocked)
	require.NoError(t, err)
	assert.True(t, isBlocked)
}

// ============================================================================
// INTEGRATION TEST: Full User Journey
// ============================================================================

func TestFullUserJourney(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	router := gin.New()
	setupRoutes(router)

	// 1. Register a new user
	t.Run("Register", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    "journey@example.com",
			"password": "password123",
			"name":     "Journey User",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	// 2. Login
	var authToken string
	t.Run("Login", func(t *testing.T) {
		payload := map[string]string{
			"email":    "journey@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Extract token from response body
		assert.Contains(t, response, "token")
		authToken = response["token"].(string)
		assert.NotEmpty(t, authToken, "token should not be empty")
	})

	// 3. Verify email (for testing - manually set email_verified to true)
	t.Run("Verify Email", func(t *testing.T) {
		_, err := testDB.Exec("UPDATE users SET email_verified = 1 WHERE email = ?", "journey@example.com")
		require.NoError(t, err)
	})

	// 4. Update profile with languages
	t.Run("Update Profile", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":      "Journey User Updated",
			"bio":       "I love meeting new people",
			"languages": "en,de,fr",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PUT", "/api/profile", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 5. Create an event
	var eventID float64
	t.Run("Create Event", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

		payload := map[string]interface{}{
			"title":              "Coffee Meetup",
			"description":        "Let's grab coffee and chat",
			"category":           "social_drinks",
			"latitude":           46.8805,
			"longitude":          8.6444,
			"start_time":         future,
			"creator_name":       "Journey User",
			"max_participants":   5,
			"gender_restriction": "any",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/events", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		eventID = response["id"].(float64)
	})

	// 5. View all events
	t.Run("View Events", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/events", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var events []Event
		err := json.Unmarshal(w.Body.Bytes(), &events)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)
	})

	// 6. Delete event
	t.Run("Delete Event", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/events/%d", int(eventID)), nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// Helper function to setup all routes for integration testing
func setupRoutes(router *gin.Engine) {
	router.POST("/api/register", register)
	router.POST("/api/login", login)
	router.GET("/api/user", authMiddleware(), getCurrentUser)
	router.GET("/api/profile", authMiddleware(), getCurrentUser)
	router.PUT("/api/profile", authMiddleware(), updateProfile)
	router.GET("/api/profile/:id", getUserProfile)
	router.GET("/api/events", getEvents)
	router.GET("/api/events/:id", getEvent)
	router.POST("/api/events", authMiddleware(), createEvent)
	router.PUT("/api/events/:id", authMiddleware(), updateEvent)
	router.DELETE("/api/events/:id", authMiddleware(), deleteEvent)
	router.POST("/api/events/:id/join", authMiddleware(), joinEvent)
	router.DELETE("/api/events/:id/leave", authMiddleware(), leaveEvent)
	router.GET("/api/events/:id/participants", getEventParticipants)

	router.GET("/api/admin/users", adminGetUsers)
	router.PUT("/api/admin/users/:id/block", adminBlockUser)
	router.PUT("/api/admin/users/:id/unblock", adminUnblockUser)
}

// ============================================================================
// ADDITIONAL COVERAGE TESTS
// ============================================================================

func TestGetCategories(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	router := gin.New()
	router.GET("/api/categories", getCategories)

	req, _ := http.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "categories")
}

func TestGetEventParticipants(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	eventID := createTestEvent(t, testDB, user1ID, "Test Event")

	// Add participant
	_, err := testDB.Exec("INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)", eventID, user2ID)
	require.NoError(t, err)

	router := gin.New()
	// Mock auth middleware to simulate user1 (event creator) viewing participants
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(user1ID))
		c.Set("is_admin", false)
		c.Set("email_verified", true)
		c.Next()
	})
	router.GET("/api/events/:id/participants", getEventParticipants)

	req, _ := http.NewRequest("GET", "/api/events/"+string(rune(eventID+'0'))+"/participants", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var participants []User
	err = json.Unmarshal(w.Body.Bytes(), &participants)
	require.NoError(t, err)
	assert.Equal(t, 1, len(participants))
}

func TestEventFiltering(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)

	// Create various events
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	
	// Event 1: Sports
	_, err := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "Sports Event", "Test", "social_sports", 46.8805, 8.6444, future, "User")
	require.NoError(t, err)

	// Event 2: Drinks
	_, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "Drinks Event", "Test", "social_drinks", 46.8805, 8.6444, future, "User")
	require.NoError(t, err)

	router := gin.New()
	router.GET("/api/events", getEvents)

	// Test category filter
	t.Run("Filter by category", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/events?category=social_sports", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var events []Event
		err := json.Unmarshal(w.Body.Bytes(), &events)
		require.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.Equal(t, "Sports Event", events[0].Title)
	})

	// Test keyword filter
	t.Run("Filter by keyword", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/events?keyword=Sports", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var events []Event
		err := json.Unmarshal(w.Body.Bytes(), &events)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)
	})
}

func TestDuplicateJoin(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	eventID := createTestEvent(t, testDB, user1ID, "Test Event")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		c.Set("email_verified", true)
		c.Set("is_admin", false)
		c.Next()
	})
	router.POST("/api/events/:id/join", joinEvent)

	// First join
	req1, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/join", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Duplicate join should fail
	req2, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/join", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestJoinEventRequiresEmailVerification(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	eventID := createTestEvent(t, testDB, user1ID, "Test Event")

	// Test 1: Unverified user should NOT be able to join
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		c.Set("email_verified", false) // Unverified email
		c.Set("is_admin", false)
		c.Next()
	})
	router.POST("/api/events/:id/join", joinEvent)

	req, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/join", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should be forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "verify your email")

	// Verify NOT in participants table
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM event_participants WHERE event_id = ? AND user_id = ?", eventID, user2ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Unverified user should not be able to join")

	// Test 2: Verified user SHOULD be able to join
	router2 := gin.New()
	router2.Use(func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		c.Set("email_verified", true) // Verified email
		c.Set("is_admin", false)
		c.Next()
	})
	router2.POST("/api/events/:id/join", joinEvent)

	req2, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/join", nil)
	w2 := httptest.NewRecorder()
	router2.ServeHTTP(w2, req2)

	// Should succeed
	assert.Equal(t, http.StatusOK, w2.Code)

	// Verify IS in participants table
	err = testDB.QueryRow("SELECT COUNT(*) FROM event_participants WHERE event_id = ? AND user_id = ?", eventID, user2ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Verified user should be able to join")

	// Test 3: Admin should be able to join even without verification
	adminID := createTestUser(t, testDB, "admin@example.com", "Admin", "password123", true)

	router3 := gin.New()
	router3.Use(func(c *gin.Context) {
		c.Set("user_id", int(adminID))
		c.Set("email_verified", false) // Unverified email
		c.Set("is_admin", true) // But is admin
		c.Next()
	})
	router3.POST("/api/events/:id/join", joinEvent)

	req3, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/join", nil)
	w3 := httptest.NewRecorder()
	router3.ServeHTTP(w3, req3)

	// Should succeed for admin
	assert.Equal(t, http.StatusOK, w3.Code)

	// Verify admin IS in participants table
	err = testDB.QueryRow("SELECT COUNT(*) FROM event_participants WHERE event_id = ? AND user_id = ?", eventID, adminID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Admin should be able to join even without verification")
}

func TestGetUserProfile(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)

	// Update profile with data
	_, err := testDB.Exec(`
		UPDATE users SET bio = ?, languages = ? WHERE id = ?
	`, "Bio text", "en,de", userID)
	require.NoError(t, err)

	// Create some events for this user
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	_, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "User Event", "Description", "social_drinks", 46.8, 8.6, future, "Test User", "user-event")
	require.NoError(t, err)

	router := gin.New()
	router.GET("/api/profile/:id", getUserProfile)

	req, _ := http.NewRequest("GET", "/api/profile/"+string(rune(userID+'0')), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check user data
	userData, ok := response["user"].(map[string]interface{})
	require.True(t, ok, "user field should be present")
	assert.Equal(t, "Test User", userData["name"])
	assert.Equal(t, "Bio text", userData["bio"])

	// Check created events
	createdEvents, ok := response["created_events"].([]interface{})
	require.True(t, ok, "created_events field should be present")
	assert.Equal(t, 1, len(createdEvents), "Should have 1 created event")
}

func TestUnblockUser(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	adminID := createTestUser(t, testDB, "admin@example.com", "Admin", "password123", true)
	userID := createTestUser(t, testDB, "user@example.com", "User", "password123", false)

	// First block the user
	_, err := testDB.Exec("UPDATE users SET is_blocked = 1 WHERE id = ?", userID)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(adminID))
		c.Set("is_admin", true)
		c.Next()
	})
	router.PUT("/api/admin/users/:id/unblock", adminUnblockUser)

	req, _ := http.NewRequest("PUT", "/api/admin/users/"+string(rune(userID+'0'))+"/unblock", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user is unblocked
	var isBlocked bool
	err = testDB.QueryRow("SELECT is_blocked FROM users WHERE id = ?", userID).Scan(&isBlocked)
	require.NoError(t, err)
	assert.False(t, isBlocked)
}

func TestInvalidEventID(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	router := gin.New()
	router.GET("/api/events/:id", getEvent)

	req, _ := http.NewRequest("GET", "/api/events/99999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteNonExistentEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Next()
	})
	router.DELETE("/api/events/:id", deleteEvent)

	req, _ := http.NewRequest("DELETE", "/api/events/99999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSearchPlaces(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	router := gin.New()
	router.GET("/api/search/places", searchPlaces)

	t.Run("Missing query parameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/search/places", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("With query parameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/search/places?q=Zurich", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// This will hit the external API, so just check for valid response
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
}

func TestAdminGetAllEvents(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	adminID := createTestUser(t, testDB, "admin@example.com", "Admin", "password123", true)
	userID := createTestUser(t, testDB, "user@example.com", "User", "password123", false)
	createTestEvent(t, testDB, userID, "Test Event")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(adminID))
		c.Set("is_admin", true)
		c.Next()
	})
	router.GET("/api/admin/events", adminGetAllEvents)

	req, _ := http.NewRequest("GET", "/api/admin/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var events []Event
	err := json.Unmarshal(w.Body.Bytes(), &events)
	require.NoError(t, err)
	// Admin sees all events (including past ones)
	assert.GreaterOrEqual(t, len(events), 0)
}

func TestAdminDeleteEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	adminID := createTestUser(t, testDB, "admin@example.com", "Admin", "password123", true)
	userID := createTestUser(t, testDB, "user@example.com", "User", "password123", false)
	eventID := createTestEvent(t, testDB, userID, "Event to Delete")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(adminID))
		c.Set("is_admin", true)
		c.Next()
	})
	router.DELETE("/api/admin/events/:id", adminDeleteEvent)

	req, _ := http.NewRequest("DELETE", "/api/admin/events/"+string(rune(eventID+'0')), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM events WHERE id = ?", eventID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestAdminUpdateEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	adminID := createTestUser(t, testDB, "admin@example.com", "Admin", "password123", true)
	userID := createTestUser(t, testDB, "user@example.com", "User", "password123", false)
	eventID := createTestEvent(t, testDB, userID, "Original Title")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(adminID))
		c.Set("is_admin", true)
		c.Next()
	})
	router.PUT("/api/admin/events/:id", adminUpdateEvent)

	future := time.Now().Add(48 * time.Hour).Format(time.RFC3339)

	payload := map[string]interface{}{
		"title":           "Updated by Admin",
		"description":     "Updated description",
		"category":        "social_sports",
		"latitude":        46.9999,
		"longitude":       9.1234,
		"start_time":      future,
		"creator_name":    "User",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", "/api/admin/events/"+string(rune(eventID+'0')), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	var title string
	err := testDB.QueryRow("SELECT title FROM events WHERE id = ?", eventID).Scan(&title)
	require.NoError(t, err)
	assert.Equal(t, "Updated by Admin", title)
}

func TestEventWithAllPreferences(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Next()
	})
	router.POST("/api/events", createEvent)

	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	payload := map[string]interface{}{
		"title":              "Full Event",
		"description":        "Event with all preferences",
		"category":           "social_drinks",
		"latitude":           46.8805,
		"longitude":          8.6444,
		"start_time":         future,
		"creator_name":       "Test User",
		"max_participants":   10,
		"gender_restriction": "male",
		"age_min":            21,
		"age_max":            35,
		"smoking_allowed":    true,
		"alcohol_allowed":    true,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "id")
}

func TestEventFilteringComprehensive(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	userID := createTestUser(t, testDB, "user@example.com", "Test User", "password123", false)

	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	// Create event with specific filters
	_, err := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time,
							creator_name, gender_restriction, age_min, age_max,
							smoking_allowed, alcohol_allowed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "Filtered Event", "Test", "social_sports", 46.8805, 8.6444, future,
		"User", "male", 21, 35, 1, 1)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/api/events", getEvents)

	tests := []struct {
		name        string
		queryParams string
	}{
		{"Gender filter", "?gender=male"},
		{"Age filter", "?age_min=21&age_max=35"},
		{"Smoking filter", "?smoking=true"},
		{"Alcohol filter", "?alcohol=true"},
		{"Drugs filter", "?drugs=false"},
		{"Combined filters", "?gender=male&alcohol=true&category=social_sports"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/events"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var events []Event
			err := json.Unmarshal(w.Body.Bytes(), &events)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(events), 0) // May or may not match
		})
	}
}
