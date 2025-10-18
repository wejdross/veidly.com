package main

import (
	"database/sql"
	"log"
)

// ApplyPrivacyFilters applies privacy rules to an event based on viewer's status
// Returns the filtered event with sensitive information hidden based on:
// - Whether the viewer is authenticated
// - Whether the viewer's email is verified
// - Whether the viewer is a participant of the event
// - Whether the viewer is the event creator
// - Event-specific privacy settings
func ApplyPrivacyFilters(event *Event, viewerUserID int, viewerIsVerified bool, isAdmin bool) {
	// Admins and event creators see everything
	if isAdmin || event.UserID == viewerUserID {
		return // No filtering needed
	}

	// Check if viewer is a participant of this event
	isParticipant := event.IsParticipant
	if viewerUserID > 0 && !isParticipant && db != nil {
		// Double-check from database if not already set (skip if db is nil for testing)
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) FROM event_participants
			WHERE event_id = ? AND user_id = ?
		`, event.ID, viewerUserID).Scan(&count)
		if err == nil && count > 0 {
			isParticipant = true
			event.IsParticipant = true
		}
	}

	// Apply organizer privacy filter
	if event.HideOrganizerUntilJoined && !isParticipant {
		event.CreatorName = "🔒 Join to see organizer"
		event.UserEmail = ""
	} else if !viewerIsVerified {
		// Unverified users see limited organizer info
		event.UserEmail = ""
	}

	// Apply participants privacy filter
	if event.HideParticipantsUntilJoined && !isParticipant {
		// Clear participant list, only show count
		event.Participants = []User{}
	}

	// Log privacy filtering (for debugging)
	log.Printf("🔒 Privacy filter applied to event %d: viewer=%d, verified=%v, participant=%v, organizer_hidden=%v",
		event.ID, viewerUserID, viewerIsVerified, isParticipant, event.HideOrganizerUntilJoined)
}

// CheckEventViewPermission checks if a user can view an event based on privacy settings
// Returns error message if viewing is not allowed, empty string if allowed
// viewerUserID: 0 for unregistered users, >0 for registered users
// viewerIsVerified: email verification status (only relevant if viewerUserID > 0)
func CheckEventViewPermission(event *Event, viewerUserID int, viewerIsVerified bool, isAdmin bool) string {
	// Admins can always view
	if isAdmin {
		return ""
	}

	// Check if event allows unregistered users
	// If unregistered users are allowed, anyone can view
	if event.AllowUnregisteredUsers {
		return ""
	}

	// If unregistered users are NOT allowed, require authentication
	if viewerUserID == 0 {
		return "This event requires registration to view. Please create an account or log in."
	}

	// User is registered - check if event requires verified email
	if event.RequireVerifiedToView && !viewerIsVerified {
		return "This event is only visible to users with verified email addresses. Please verify your email to view event details."
	}

	return ""
}

// CheckEventJoinPermission checks if a user can join an event based on privacy settings
// Returns error message if joining is not allowed, empty string if allowed
func CheckEventJoinPermission(event *Event, viewerIsVerified bool, isAdmin bool) string {
	// Admins can always join (for moderation purposes)
	if isAdmin {
		return ""
	}

	// Check if event requires verified email to join
	if event.RequireVerifiedToJoin && !viewerIsVerified {
		return "This event requires a verified email address to join. Please verify your email first."
	}

	return ""
}

// GetParticipantsWithPrivacy retrieves event participants with privacy filtering
func GetParticipantsWithPrivacy(eventID int, viewerUserID int, viewerIsVerified bool, isAdmin bool) ([]User, error) {
	// First get the event to check privacy settings
	var hideParticipants bool
	var creatorID int
	var isParticipant bool

	err := db.QueryRow(`
		SELECT user_id, hide_participants_until_joined,
		       EXISTS(SELECT 1 FROM event_participants WHERE event_id = ? AND user_id = ?) as is_participant
		FROM events WHERE id = ?
	`, eventID, viewerUserID, eventID).Scan(&creatorID, &hideParticipants, &isParticipant)

	if err != nil {
		return nil, err
	}

	// Admins, creators, and participants can always see the list
	if isAdmin || creatorID == viewerUserID || isParticipant {
		return getFullParticipantList(eventID)
	}

	// If participants are hidden and viewer is not a participant, return empty list
	if hideParticipants {
		log.Printf("🔒 Participant list hidden for event %d: viewer %d is not a participant", eventID, viewerUserID)
		return []User{}, nil
	}

	// Otherwise return the full list (but maybe with limited info for unverified users)
	participants, err := getFullParticipantList(eventID)
	if err != nil {
		return nil, err
	}

	// If viewer is not verified, hide contact information
	if !viewerIsVerified {
		for i := range participants {
			participants[i].Email = ""
		}
	}

	return participants, nil
}

// getFullParticipantList retrieves all participants for an event (internal helper)
func getFullParticipantList(eventID int) ([]User, error) {
	rows, err := db.Query(`
		SELECT u.id, u.name, u.email, u.bio, u.languages, ep.joined_at
		FROM event_participants ep
		JOIN users u ON ep.user_id = u.id
		WHERE ep.event_id = ?
		ORDER BY ep.joined_at ASC
	`, eventID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []User
	for rows.Next() {
		var u User
		var bio, languages sql.NullString
		var joinedAt sql.NullTime
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &bio, &languages, &joinedAt)
		if err != nil {
			continue
		}
		if bio.Valid {
			u.Bio = bio.String
		}
		if languages.Valid {
			u.Languages = languages.String
		}
		participants = append(participants, u)
	}

	return participants, nil
}
