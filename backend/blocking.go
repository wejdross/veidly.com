package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// blockUser blocks a user (POST /api/users/:id/block)
func blockUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	blockedIDStr := c.Param("id")
	blockedID, err := strconv.Atoi(blockedIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	blockerID := userID.(int)

	// Prevent self-blocking
	if blockerID == blockedID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot block yourself"})
		return
	}

	// Check if blocked user exists
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE id = ?`, blockedID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Reason is optional, so ignore binding errors
		req.Reason = ""
	}

	// Insert block (will fail if already blocked due to UNIQUE constraint)
	_, err = db.Exec(`
		INSERT INTO user_blocks (blocker_id, blocked_id, reason)
		VALUES (?, ?, ?)
	`, blockerID, blockedID, req.Reason)

	if err != nil {
		// Check if it's a duplicate key error
		if err.Error() == "UNIQUE constraint failed: user_blocks.blocker_id, user_blocks.blocked_id" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already blocked"})
			return
		}
		log.Printf("❌ Error blocking user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to block user"})
		return
	}

	log.Printf("🚫 User %d blocked user %d", blockerID, blockedID)
	c.JSON(http.StatusOK, gin.H{"message": "User blocked successfully"})
}

// unblockUser unblocks a user (DELETE /api/users/:id/block)
func unblockUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	blockedIDStr := c.Param("id")
	blockedID, err := strconv.Atoi(blockedIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	blockerID := userID.(int)

	// Delete the block
	result, err := db.Exec(`
		DELETE FROM user_blocks
		WHERE blocker_id = ? AND blocked_id = ?
	`, blockerID, blockedID)

	if err != nil {
		log.Printf("❌ Error unblocking user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock user"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Block not found"})
		return
	}

	log.Printf("✅ User %d unblocked user %d", blockerID, blockedID)
	c.JSON(http.StatusOK, gin.H{"message": "User unblocked successfully"})
}

// getBlockedUsers returns list of users that the current user has blocked (GET /api/blocks)
func getBlockedUsers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	blockerID := userID.(int)

	rows, err := db.Query(`
		SELECT ub.id, ub.blocked_id, u.name, u.email, ub.reason, ub.created_at
		FROM user_blocks ub
		JOIN users u ON ub.blocked_id = u.id
		WHERE ub.blocker_id = ?
		ORDER BY ub.created_at DESC
	`, blockerID)

	if err != nil {
		log.Printf("❌ Error fetching blocked users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blocked users"})
		return
	}
	defer rows.Close()

	var blocks []map[string]interface{}
	for rows.Next() {
		var blockID, blockedID int
		var name, email string
		var reason sql.NullString
		var createdAt string

		err := rows.Scan(&blockID, &blockedID, &name, &email, &reason, &createdAt)
		if err != nil {
			log.Printf("❌ Error scanning blocked user: %v", err)
			continue
		}

		block := map[string]interface{}{
			"id":         blockID,
			"blocked_id": blockedID,
			"name":       name,
			"email":      email,
			"created_at": createdAt,
		}

		if reason.Valid {
			block["reason"] = reason.String
		}

		blocks = append(blocks, block)
	}

	if blocks == nil {
		blocks = []map[string]interface{}{}
	}

	log.Printf("📋 User %d fetched %d blocked users", blockerID, len(blocks))
	c.JSON(http.StatusOK, blocks)
}

// AreUsersBlocked checks if two users have a block relationship (either direction)
func AreUsersBlocked(userID1, userID2 int) bool {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM user_blocks
		WHERE (blocker_id = ? AND blocked_id = ?)
		   OR (blocker_id = ? AND blocked_id = ?)
	`, userID1, userID2, userID2, userID1).Scan(&count)

	return err == nil && count > 0
}

// FilterEventsByBlocks filters out events from blocked users
func FilterEventsByBlocks(events []Event, userID int) []Event {
	if userID == 0 {
		return events
	}

	// Get all blocked user IDs (both directions)
	rows, err := db.Query(`
		SELECT blocked_id FROM user_blocks WHERE blocker_id = ?
		UNION
		SELECT blocker_id FROM user_blocks WHERE blocked_id = ?
	`, userID, userID)

	if err != nil {
		log.Printf("⚠️  Error fetching blocks for filtering: %v", err)
		return events
	}
	defer rows.Close()

	blockedUserIDs := make(map[int]bool)
	for rows.Next() {
		var blockedID int
		if err := rows.Scan(&blockedID); err == nil {
			blockedUserIDs[blockedID] = true
		}
	}

	// Filter events
	filtered := []Event{}
	for _, event := range events {
		if !blockedUserIDs[event.UserID] {
			filtered = append(filtered, event)
		}
	}

	if len(filtered) < len(events) {
		log.Printf("🔒 Filtered %d events from blocked users for user %d", len(events)-len(filtered), userID)
	}

	return filtered
}
