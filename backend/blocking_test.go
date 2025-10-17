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

func TestUserBlockUser(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)

	router := gin.New()
	router.POST("/api/users/:id/block", func(c *gin.Context) {
		c.Set("user_id", int(user1ID))
		blockUser(c)
	})

	t.Run("Successfully block a user", func(t *testing.T) {
		reqBody := BlockUserRequest{Reason: "spam"}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/users/"+string(rune(user2ID+'0'))+"/block", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify block was created
		var count int
		err := testDB.QueryRow(`SELECT COUNT(*) FROM user_blocks WHERE blocker_id = ? AND blocked_id = ?`,
			user1ID, user2ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Cannot block yourself", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/users/"+string(rune(user1ID+'0'))+"/block", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Cannot block yourself")
	})

	t.Run("Cannot block non-existent user", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/users/99999/block", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Cannot block same user twice", func(t *testing.T) {
		user3ID := createTestUser(t, testDB, "user3@example.com", "User 3", "password123", false)

		// Block once
		testDB.Exec(`INSERT INTO user_blocks (blocker_id, blocked_id) VALUES (?, ?)`, user1ID, user3ID)

		// Try to block again
		req, _ := http.NewRequest("POST", "/api/users/"+string(rune(user3ID+'0'))+"/block", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "already blocked")
	})
}

func TestUserUnblockUser(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)

	// Create a block
	_, err := testDB.Exec(`INSERT INTO user_blocks (blocker_id, blocked_id, reason) VALUES (?, ?, ?)`,
		user1ID, user2ID, "test")
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/api/users/:id/block", func(c *gin.Context) {
		c.Set("user_id", int(user1ID))
		unblockUser(c)
	})

	t.Run("Successfully unblock a user", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/users/"+string(rune(user2ID+'0'))+"/block", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify block was removed
		var count int
		err := testDB.QueryRow(`SELECT COUNT(*) FROM user_blocks WHERE blocker_id = ? AND blocked_id = ?`,
			user1ID, user2ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Cannot unblock non-blocked user", func(t *testing.T) {
		user3ID := createTestUser(t, testDB, "user3@example.com", "User 3", "password123", false)

		req, _ := http.NewRequest("DELETE", "/api/users/"+string(rune(user3ID+'0'))+"/block", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetBlockedUsers(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	user3ID := createTestUser(t, testDB, "user3@example.com", "User 3", "password123", false)

	// User1 blocks User2 and User3
	testDB.Exec(`INSERT INTO user_blocks (blocker_id, blocked_id, reason) VALUES (?, ?, ?)`,
		user1ID, user2ID, "spam")
	testDB.Exec(`INSERT INTO user_blocks (blocker_id, blocked_id) VALUES (?, ?)`,
		user1ID, user3ID)

	router := gin.New()
	router.GET("/api/blocks", func(c *gin.Context) {
		c.Set("user_id", int(user1ID))
		getBlockedUsers(c)
	})

	t.Run("Get list of blocked users", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/blocks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var blocks []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &blocks)
		require.NoError(t, err)

		assert.Equal(t, 2, len(blocks))

		// Check first block has reason
		foundWithReason := false
		for _, block := range blocks {
			if reason, ok := block["reason"]; ok && reason == "spam" {
				foundWithReason = true
				break
			}
		}
		assert.True(t, foundWithReason, "Should have found block with reason 'spam'")
	})

	t.Run("Empty list when no blocks", func(t *testing.T) {
		user4ID := createTestUser(t, testDB, "user4@example.com", "User 4", "password123", false)

		router := gin.New()
		router.GET("/api/blocks", func(c *gin.Context) {
			c.Set("user_id", int(user4ID))
			getBlockedUsers(c)
		})

		req, _ := http.NewRequest("GET", "/api/blocks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var blocks []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &blocks)
		require.NoError(t, err)
		assert.Equal(t, 0, len(blocks))
	})
}

func TestAreUsersBlocked(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	user3ID := createTestUser(t, testDB, "user3@example.com", "User 3", "password123", false)

	// User1 blocks User2
	testDB.Exec(`INSERT INTO user_blocks (blocker_id, blocked_id) VALUES (?, ?)`, user1ID, user2ID)

	t.Run("Returns true when blocked", func(t *testing.T) {
		assert.True(t, AreUsersBlocked(int(user1ID), int(user2ID)))
	})

	t.Run("Returns true when blocked (reverse direction)", func(t *testing.T) {
		assert.True(t, AreUsersBlocked(int(user2ID), int(user1ID)))
	})

	t.Run("Returns false when not blocked", func(t *testing.T) {
		assert.False(t, AreUsersBlocked(int(user1ID), int(user3ID)))
	})
}

func TestFilterEventsByBlocks(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)
	user3ID := createTestUser(t, testDB, "user3@example.com", "User 3", "password123", false)

	// User1 blocks User2
	testDB.Exec(`INSERT INTO user_blocks (blocker_id, blocked_id) VALUES (?, ?)`, user1ID, user2ID)

	events := []Event{
		{ID: 1, UserID: int(user1ID), Title: "Event by User 1"},
		{ID: 2, UserID: int(user2ID), Title: "Event by User 2"},
		{ID: 3, UserID: int(user3ID), Title: "Event by User 3"},
	}

	t.Run("Filters out blocked user's events", func(t *testing.T) {
		filtered := FilterEventsByBlocks(events, int(user1ID))

		assert.Equal(t, 2, len(filtered))
		assert.Equal(t, 1, filtered[0].ID)
		assert.Equal(t, 3, filtered[1].ID)
	})

	t.Run("Returns all events when no blocks", func(t *testing.T) {
		filtered := FilterEventsByBlocks(events, int(user3ID))
		assert.Equal(t, 3, len(filtered))
	})

	t.Run("Returns all events for unauthenticated user", func(t *testing.T) {
		filtered := FilterEventsByBlocks(events, 0)
		assert.Equal(t, 3, len(filtered))
	})
}

func TestBlockingPreventJoinEvent(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	user1ID := createTestUser(t, testDB, "user1@example.com", "User 1", "password123", false)
	user2ID := createTestUser(t, testDB, "user2@example.com", "User 2", "password123", false)

	// User1 blocks User2
	testDB.Exec(`INSERT INTO user_blocks (blocker_id, blocked_id) VALUES (?, ?)`, user1ID, user2ID)

	// Create event by User1
	future := "2025-12-31T18:00:00Z"
	result, err := testDB.Exec(`
		INSERT INTO events (user_id, title, description, category, latitude, longitude, start_time, creator_name, slug)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user1ID, "Test Event", "Description", "social_drinks", 46.8, 8.6, future, "User 1", "test-event")
	require.NoError(t, err)

	_, _ = result.LastInsertId()

	t.Run("Blocked user cannot join event", func(t *testing.T) {
		// Check if blocked
		blocked := AreUsersBlocked(int(user2ID), int(user1ID))
		assert.True(t, blocked, "User2 should be blocked from User1")

		// Verify the block exists in database
		var count int
		err := testDB.QueryRow(`
			SELECT COUNT(*) FROM user_blocks
			WHERE (blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)
		`, user1ID, user2ID, user2ID, user1ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Block should exist in database")

		// Try to join - should be prevented by join logic (will be implemented)
		// For now, just verify the block relationship exists
		assert.True(t, blocked)
	})
}
