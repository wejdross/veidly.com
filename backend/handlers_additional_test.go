package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetPublicEvent tests the public event endpoint
func TestGetPublicEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)

	db = testDB

	// Create a test user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "user@example.com", hashedPassword, "Test User", 1)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	// Create a test event
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	_, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time,
							creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "Public Test Event", "Test Description", "social_drinks", 52.52, 13.405, future,
		"Test User", "public-test-event")
	require.NoError(t, err)

	router := gin.New()
	router.GET("/api/public/events/:slug", getPublicEvent)

	// Test getting public event
	req, _ := http.NewRequest("GET", "/api/public/events/public-test-event", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var event Event
	err = json.Unmarshal(w.Body.Bytes(), &event)
	require.NoError(t, err)
	assert.Equal(t, "Public Test Event", event.Title)
	assert.Equal(t, "public-test-event", event.Slug)
}

// TestGetPublicEventNotFound tests 404 for non-existent event
func TestGetPublicEventNotFound(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)

	db = testDB

	router := gin.New()
	router.GET("/api/public/events/:slug", getPublicEvent)

	req, _ := http.NewRequest("GET", "/api/public/events/non-existent-slug", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestLeaveEventComprehensive tests leaving events
func TestLeaveEventComprehensive(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)

	db = testDB

	// Create test users
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "organizer@example.com", hashedPassword, "Organizer", 1)
	require.NoError(t, err)
	organizerID, _ := result.LastInsertId()

	result, err = testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "participant@example.com", hashedPassword, "Participant", 1)
	require.NoError(t, err)
	participantID, _ := result.LastInsertId()

	// Create event
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	result, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time,
							creator_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, organizerID, "Leave Test Event", "Description", "social_drinks", 52.52, 13.405, future,
		"Organizer")
	require.NoError(t, err)
	eventID, _ := result.LastInsertId()

	// Add participant
	_, err = testDB.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`, eventID, participantID)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(participantID))
		c.Set("email_verified", true)
		c.Set("is_admin", false)
		c.Next()
	})
	router.DELETE("/api/events/:id/leave", leaveEvent)

	// Test leaving event
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/events/%d/leave", eventID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify participant was removed
	var count int
	err = testDB.QueryRow(`SELECT COUNT(*) FROM event_participants WHERE event_id = ? AND user_id = ?`,
		eventID, participantID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestLeaveEventNotParticipant tests leaving event when not a participant
func TestLeaveEventNotParticipant(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)

	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "user@example.com", hashedPassword, "User", 1)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	// Create event
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	result, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time,
							creator_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "Event", "Description", "social_drinks", 52.52, 13.405, future,
		"User")
	require.NoError(t, err)
	eventID, _ := result.LastInsertId()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Set("email_verified", true)
		c.Set("is_admin", false)
		c.Next()
	})
	router.DELETE("/api/events/:id/leave", leaveEvent)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/events/%d/leave", eventID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDownloadEventICS tests ICS calendar file download
func TestDownloadEventICS(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)

	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "user@example.com", hashedPassword, "Test User", 1)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	// Create event with specific times
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	result, err = testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, end_time,
							creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, "ICS Test Event", "Test Description", "social_drinks", 52.52, 13.405,
		startTime.Format(time.RFC3339), endTime.Format(time.RFC3339),
		"Test User", "ics-test-event")
	require.NoError(t, err)

	router := gin.New()
	router.GET("/api/public/events/:slug/ics", downloadEventICS)

	req, _ := http.NewRequest("GET", "/api/public/events/ics-test-event/ics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/calendar")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "ics-test-event.ics")
	assert.Contains(t, w.Body.String(), "BEGIN:VCALENDAR")
	assert.Contains(t, w.Body.String(), "ICS Test Event")
	assert.Contains(t, w.Body.String(), "END:VCALENDAR")
}

// TestSearchPlacesIntegration tests the place search functionality
func TestSearchPlacesIntegration(t *testing.T) {
	router := gin.New()
	router.GET("/api/search-places", searchPlaces)

	// Test valid search
	req, _ := http.NewRequest("GET", "/api/search-places?q=Warsaw", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var places []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &places)
	require.NoError(t, err)
	// Note: This might fail if no internet connection, but tests real API
}

// TestSearchPlacesEmptyQuery tests search with empty query
func TestSearchPlacesEmptyQuery(t *testing.T) {
	router := gin.New()
	router.GET("/api/search-places", searchPlaces)

	req, _ := http.NewRequest("GET", "/api/search-places?q=", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAdminVerifyUserEmail tests admin email verification endpoint
func TestAdminVerifyUserEmail(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)

	db = testDB

	// Create admin user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified, is_admin)
		VALUES (?, ?, ?, ?, ?)
	`, "admin@example.com", hashedPassword, "Admin", 1, 1)
	require.NoError(t, err)
	adminID, _ := result.LastInsertId()

	// Create unverified user
	result, err = testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "unverified@example.com", hashedPassword, "Unverified", 0)
	require.NoError(t, err)
	unverifiedID, _ := result.LastInsertId()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(adminID))
		c.Set("email_verified", true)
		c.Set("is_admin", true)
		c.Next()
	})
	router.PUT("/api/admin/users/:id/verify-email", adminVerifyUserEmail)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/admin/users/%d/verify-email", unverifiedID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user is now verified
	var emailVerified bool
	err = testDB.QueryRow(`SELECT email_verified FROM users WHERE id = ?`, unverifiedID).Scan(&emailVerified)
	require.NoError(t, err)
	assert.True(t, emailVerified)
}

// TestCreateEventValidationErrors tests various validation errors
func TestCreateEventValidationErrors(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)

	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "user@example.com", hashedPassword, "User", 1)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Set("email_verified", true)
		c.Set("is_admin", false)
		c.Next()
	})
	router.POST("/api/events", createEvent)

	tests := []struct {
		name       string
		payload    map[string]interface{}
		expectCode int
	}{
		{
			name: "Missing title",
			payload: map[string]interface{}{
				"description":      "Test",
				"category":         "social_drinks",
				"latitude":         52.52,
				"longitude":        13.405,
				"start_time":       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"creator_name":     "User","gender_restriction": "any",
				"age_min":          18,
				"age_max":          99,
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Title too short",
			payload: map[string]interface{}{
				"title":            "Hi",
				"description":      "Test description",
				"category":         "social_drinks",
				"latitude":         52.52,
				"longitude":        13.405,
				"start_time":       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"creator_name":     "User","gender_restriction": "any",
				"age_min":          18,
				"age_max":          99,
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Invalid category",
			payload: map[string]interface{}{
				"title":            "Valid Title",
				"description":      "Test description",
				"category":         "invalid_category",
				"latitude":         52.52,
				"longitude":        13.405,
				"start_time":       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"creator_name":     "User","gender_restriction": "any",
				"age_min":          18,
				"age_max":          99,
			},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/events", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}

// TestUpdateProfileValidation tests profile update validation
func TestUpdateProfileValidation(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)

	db = testDB

	// Create test user
	hashedPassword, _ := hashPassword("password123")
	result, err := testDB.Exec(`
		INSERT INTO users (email, password, name, email_verified)
		VALUES (?, ?, ?, ?)
	`, "user@example.com", hashedPassword, "User", 1)
	require.NoError(t, err)
	userID, _ := result.LastInsertId()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int(userID))
		c.Set("email_verified", true)
		c.Set("is_admin", false)
		c.Next()
	})
	router.PUT("/api/profile", updateProfile)

	// Test updating with valid data
	payload := map[string]interface{}{
		"name":      "Updated Name",
		"bio":       "Updated bio",
		"languages": "en,de,pl",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", "/api/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	var name, bio, languages string
	err = testDB.QueryRow(`SELECT name, bio, languages FROM users WHERE id = ?`, userID).Scan(&name, &bio, &languages)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", name)
	assert.Equal(t, "Updated bio", bio)
	assert.Equal(t, "en,de,pl", languages)
}

