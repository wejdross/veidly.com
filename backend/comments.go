package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// getEventComments retrieves all comments for an event (GET /api/events/:id/comments)
// Only accessible to event participants
func getEventComments(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	viewerID := userID.(int)

	// Check if event exists and user is a participant or creator
	var eventCreatorID int
	var isParticipant bool
	err = db.QueryRow(`
		SELECT e.user_id,
		       EXISTS(SELECT 1 FROM event_participants WHERE event_id = ? AND user_id = ?) as is_participant
		FROM events e
		WHERE e.id = ?
	`, eventID, viewerID, eventID).Scan(&eventCreatorID, &isParticipant)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
		log.Printf("❌ Error checking event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve comments"})
		return
	}

	// Only participants and creator can view comments
	isCreator := eventCreatorID == viewerID
	if !isParticipant && !isCreator {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only event participants can view comments"})
		return
	}

	// Retrieve comments (excluding soft-deleted)
	rows, err := db.Query(`
		SELECT c.id, c.event_id, c.user_id, c.comment, c.created_at, c.updated_at, u.name
		FROM event_comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.event_id = ? AND c.is_deleted = 0
		ORDER BY c.created_at ASC
	`, eventID)

	if err != nil {
		log.Printf("❌ Error fetching comments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve comments"})
		return
	}
	defer rows.Close()

	var comments []EventComment
	for rows.Next() {
		var comment EventComment
		var updatedAt sql.NullTime

		err := rows.Scan(
			&comment.ID,
			&comment.EventID,
			&comment.UserID,
			&comment.Comment,
			&comment.CreatedAt,
			&updatedAt,
			&comment.UserName,
		)
		if err != nil {
			log.Printf("❌ Error scanning comment: %v", err)
			continue
		}

		if updatedAt.Valid {
			comment.UpdatedAt = updatedAt.Time
		}

		// Mark if this comment belongs to the viewer
		comment.IsOwn = comment.UserID == viewerID

		comments = append(comments, comment)
	}

	if comments == nil {
		comments = []EventComment{}
	}

	log.Printf("💬 User %d fetched %d comments for event %d", viewerID, len(comments), eventID)
	c.JSON(http.StatusOK, comments)
}

// createEventComment creates a new comment on an event (POST /api/events/:id/comments)
// Only accessible to event participants
func createEventComment(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	viewerID := userID.(int)

	// Check if event exists and user is a participant or creator
	var eventCreatorID int
	var isParticipant bool
	var commentsEnabled bool
	err = db.QueryRow(`
		SELECT e.user_id, e.comments_enabled,
		       EXISTS(SELECT 1 FROM event_participants WHERE event_id = ? AND user_id = ?) as is_participant
		FROM events e
		WHERE e.id = ?
	`, eventID, viewerID, eventID).Scan(&eventCreatorID, &commentsEnabled, &isParticipant)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err != nil {
		log.Printf("❌ Error checking event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	// Check if comments are enabled for this event
	if !commentsEnabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Comments are disabled for this event"})
		return
	}

	// Only participants and creator can comment
	isCreator := eventCreatorID == viewerID
	if !isParticipant && !isCreator {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only event participants can comment"})
		return
	}

	// Parse request body
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Insert comment
	result, err := db.Exec(`
		INSERT INTO event_comments (event_id, user_id, comment)
		VALUES (?, ?, ?)
	`, eventID, viewerID, req.Comment)

	if err != nil {
		log.Printf("❌ Error creating comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	commentID, _ := result.LastInsertId()

	// Retrieve the created comment with user info
	var comment EventComment
	err = db.QueryRow(`
		SELECT c.id, c.event_id, c.user_id, c.comment, c.created_at, u.name
		FROM event_comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`, commentID).Scan(
		&comment.ID,
		&comment.EventID,
		&comment.UserID,
		&comment.Comment,
		&comment.CreatedAt,
		&comment.UserName,
	)

	if err != nil {
		log.Printf("⚠️  Comment created but failed to retrieve: %v", err)
		c.JSON(http.StatusCreated, gin.H{"id": commentID, "message": "Comment created"})
		return
	}

	comment.IsOwn = true

	log.Printf("💬 User %d created comment on event %d", viewerID, eventID)
	c.JSON(http.StatusCreated, comment)
}

// updateEventComment updates a comment (PUT /api/comments/:id)
// Only the comment author can update
func updateEventComment(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	viewerID := userID.(int)

	// Check if comment exists and belongs to user
	var commentUserID int
	err = db.QueryRow(`SELECT user_id FROM event_comments WHERE id = ? AND is_deleted = 0`, commentID).Scan(&commentUserID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	if err != nil {
		log.Printf("❌ Error checking comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	// Only comment author can update
	if commentUserID != viewerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own comments"})
		return
	}

	// Parse request body
	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Update comment
	_, err = db.Exec(`
		UPDATE event_comments
		SET comment = ?, updated_at = ?
		WHERE id = ?
	`, req.Comment, time.Now(), commentID)

	if err != nil {
		log.Printf("❌ Error updating comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	log.Printf("✏️  User %d updated comment %d", viewerID, commentID)
	c.JSON(http.StatusOK, gin.H{"message": "Comment updated successfully"})
}

// deleteEventComment soft-deletes a comment (DELETE /api/comments/:id)
// Only the comment author can delete
func deleteEventComment(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	viewerID := userID.(int)

	// Check if comment exists and belongs to user
	var commentUserID int
	err = db.QueryRow(`SELECT user_id FROM event_comments WHERE id = ? AND is_deleted = 0`, commentID).Scan(&commentUserID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	if err != nil {
		log.Printf("❌ Error checking comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	// Only comment author can delete
	if commentUserID != viewerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own comments"})
		return
	}

	// Soft delete comment
	_, err = db.Exec(`
		UPDATE event_comments
		SET is_deleted = 1
		WHERE id = ?
	`, commentID)

	if err != nil {
		log.Printf("❌ Error deleting comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	log.Printf("🗑️  User %d deleted comment %d", viewerID, commentID)
	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}
