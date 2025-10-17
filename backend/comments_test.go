package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEventComments(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	user3ID := createTestUser(t, testDB, "user3@example.com", "User 3", "password123", false)

	// Create event by User1
	future := "2025-12-31T18:00:00Z"
	result, err := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug, comments_enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user1ID, "Test Event", "Description", "social_drinks", 46.8, 8.6, future, "User 1", "test-event", true)
	require.NoError(t, err)
	eventID, _ := result.LastInsertId()

	// User2 joins event
	_, err = testDB.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`, eventID, user2ID)
	require.NoError(t, err)

	// Create some comments
	testDB.Exec(`INSERT INTO event_comments (event_id, user_id, comment) VALUES (?, ?, ?)`,
		eventID, user1ID, "Comment by creator")
	testDB.Exec(`INSERT INTO event_comments (event_id, user_id, comment) VALUES (?, ?, ?)`,
		eventID, user2ID, "Comment by participant")

	router := gin.New()
	router.GET("/api/events/:id/comments", func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		getEventComments(c)
	})

	t.Run("Participant can view comments", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/events/"+string(rune(eventID+'0'))+"/comments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var comments []EventComment
		err := json.Unmarshal(w.Body.Bytes(), &comments)
		require.NoError(t, err)
		assert.Equal(t, 2, len(comments))

		// Check comment structure
		assert.Equal(t, "Comment by creator", comments[0].Comment)
		assert.Equal(t, "User 1", comments[0].UserName)
		assert.False(t, comments[0].IsOwn) // Not user2's comment

		assert.Equal(t, "Comment by participant", comments[1].Comment)
		assert.True(t, comments[1].IsOwn) // User2's comment
	})

	t.Run("Non-participant cannot view comments", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/events/:id/comments", func(c *gin.Context) {
			c.Set("user_id", int(user3ID))
			getEventComments(c)
		})

		req, _ := http.NewRequest("GET", "/api/events/"+string(rune(eventID+'0'))+"/comments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "Only event participants can view comments")
	})

	t.Run("Event creator can view comments without joining", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/events/:id/comments", func(c *gin.Context) {
			c.Set("user_id", int(user1ID))
			getEventComments(c)
		})

		req, _ := http.NewRequest("GET", "/api/events/"+string(rune(eventID+'0'))+"/comments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCreateEventComment(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)

	// Create event
	future := "2025-12-31T18:00:00Z"
	result, err := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug, comments_enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user1ID, "Test Event", "Description", "social_drinks", 46.8, 8.6, future, "User 1", "test-event", true)
	require.NoError(t, err)
	eventID, _ := result.LastInsertId()

	// User2 joins event
	_, err = testDB.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`, eventID, user2ID)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/api/events/:id/comments", func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		createEventComment(c)
	})

	t.Run("Participant can create comment", func(t *testing.T) {
		reqBody := CreateCommentRequest{Comment: "This is a test comment"}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/comments", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var comment EventComment
		err := json.Unmarshal(w.Body.Bytes(), &comment)
		require.NoError(t, err)
		assert.Equal(t, "This is a test comment", comment.Comment)
		assert.Equal(t, "User 2", comment.UserName)
		assert.True(t, comment.IsOwn)

		// Verify in database
		var count int
		testDB.QueryRow(`SELECT COUNT(*) FROM event_comments WHERE event_id = ?`, eventID).Scan(&count)
		assert.Equal(t, 1, count)
	})

	t.Run("Non-participant cannot create comment", func(t *testing.T) {
		user3ID := createTestUser(t, testDB, "user3@example.com", "User 3", "password123", false)

		router := gin.New()
		router.POST("/api/events/:id/comments", func(c *gin.Context) {
			c.Set("user_id", int(user3ID))
			createEventComment(c)
		})

		reqBody := CreateCommentRequest{Comment: "Should fail"}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/comments", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Cannot create comment when comments disabled", func(t *testing.T) {
		// Create event with comments disabled
		result2, _ := testDB.Exec(`
			INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug, comments_enabled)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, user1ID, "No Comments Event", "Description", "social_drinks", 46.8, 8.6, future, "User 1", "no-comments", false)
		eventID2, _ := result2.LastInsertId()

		// User2 joins
		testDB.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`, eventID2, user2ID)

		reqBody := CreateCommentRequest{Comment: "Should fail"}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID2+'0'))+"/comments", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "Comments are disabled")
	})

	t.Run("Validation fails for empty comment", func(t *testing.T) {
		reqBody := CreateCommentRequest{Comment: ""}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/events/"+string(rune(eventID+'0'))+"/comments", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateEventComment(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)

	// Create event and join
	future := "2025-12-31T18:00:00Z"
	result, _ := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user1ID, "Test Event", "Description", "social_drinks", 46.8, 8.6, future, "User 1", "test-event")
	eventID, _ := result.LastInsertId()

	testDB.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`, eventID, user2ID)

	// Create comment by User2
	result, _ = testDB.Exec(`INSERT INTO event_comments (event_id, user_id, comment) VALUES (?, ?, ?)`,
		eventID, user2ID, "Original comment")
	commentID, _ := result.LastInsertId()

	router := gin.New()
	router.PUT("/api/comments/:id", func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		updateEventComment(c)
	})

	t.Run("User can update own comment", func(t *testing.T) {
		reqBody := UpdateCommentRequest{Comment: "Updated comment"}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/comments/"+string(rune(commentID+'0')), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in database
		var updatedComment string
		testDB.QueryRow(`SELECT comment FROM event_comments WHERE id = ?`, commentID).Scan(&updatedComment)
		assert.Equal(t, "Updated comment", updatedComment)
	})

	t.Run("User cannot update others' comments", func(t *testing.T) {
		router := gin.New()
		router.PUT("/api/comments/:id", func(c *gin.Context) {
			c.Set("user_id", int(user1ID)) // Different user
			updateEventComment(c)
		})

		reqBody := UpdateCommentRequest{Comment: "Should fail"}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/comments/"+string(rune(commentID+'0')), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestDeleteEventComment(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)

	// Create event and join
	future := "2025-12-31T18:00:00Z"
	result, _ := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user1ID, "Test Event", "Description", "social_drinks", 46.8, 8.6, future, "User 1", "test-event")
	eventID, _ := result.LastInsertId()

	testDB.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`, eventID, user2ID)

	// Create comment by User2
	result, _ = testDB.Exec(`INSERT INTO event_comments (event_id, user_id, comment) VALUES (?, ?, ?)`,
		eventID, user2ID, "Comment to delete")
	commentID, _ := result.LastInsertId()

	router := gin.New()
	router.DELETE("/api/comments/:id", func(c *gin.Context) {
		c.Set("user_id", int(user2ID))
		deleteEventComment(c)
	})

	t.Run("User can delete own comment", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/comments/"+string(rune(commentID+'0')), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify soft delete in database
		var isDeleted bool
		testDB.QueryRow(`SELECT is_deleted FROM event_comments WHERE id = ?`, commentID).Scan(&isDeleted)
		assert.True(t, isDeleted)
	})

	t.Run("User cannot delete others' comments", func(t *testing.T) {
		// Create another comment
		result, _ := testDB.Exec(`INSERT INTO event_comments (event_id, user_id, comment) VALUES (?, ?, ?)`,
			eventID, user1ID, "Comment by User1")
		comment2ID, _ := result.LastInsertId()

		router := gin.New()
		router.DELETE("/api/comments/:id", func(c *gin.Context) {
			c.Set("user_id", int(user2ID)) // User2 trying to delete User1's comment
			deleteEventComment(c)
		})

		req, _ := http.NewRequest("DELETE", "/api/comments/"+string(rune(comment2ID+'0')), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestCommentsSoftDelete(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)

	// Create event
	future := "2025-12-31T18:00:00Z"
	result, _ := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user1ID, "Test Event", "Description", "social_drinks", 46.8, 8.6, future, "User 1", "test-event")
	eventID, _ := result.LastInsertId()

	// Create and soft-delete a comment
	result, _ = testDB.Exec(`INSERT INTO event_comments (event_id, user_id, comment, is_deleted) VALUES (?, ?, ?, ?)`,
		eventID, user1ID, "Deleted comment", true)

	// Create active comment
	testDB.Exec(`INSERT INTO event_comments (event_id, user_id, comment) VALUES (?, ?, ?)`,
		eventID, user1ID, "Active comment")

	router := gin.New()
	router.GET("/api/events/:id/comments", func(c *gin.Context) {
		c.Set("user_id", int(user1ID))
		getEventComments(c)
	})

	t.Run("Soft-deleted comments are not returned", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/events/"+string(rune(eventID+'0'))+"/comments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var comments []EventComment
		json.Unmarshal(w.Body.Bytes(), &comments)
		assert.Equal(t, 1, len(comments))
		assert.Equal(t, "Active comment", comments[0].Comment)
	})
}
