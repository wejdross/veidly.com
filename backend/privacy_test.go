package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckEventJoinPermission(t *testing.T) {
	// Test 1: Unverified user cannot join event requiring verification
	event := &Event{
		ID:                    1,
		RequireVerifiedToJoin: true,
	}
	reason := CheckEventJoinPermission(event, false, false)
	assert.NotEmpty(t, reason)
	assert.Contains(t, reason, "verified email")

	// Test 2: Verified user can join
	reason = CheckEventJoinPermission(event, true, false)
	assert.Empty(t, reason)

	// Test 3: Admin can always join
	reason = CheckEventJoinPermission(event, false, true)
	assert.Empty(t, reason)

	// Test 4: Event not requiring verification
	event2 := &Event{
		ID:                    2,
		RequireVerifiedToJoin: false,
	}
	reason = CheckEventJoinPermission(event2, false, false)
	assert.Empty(t, reason)
}

func TestGetParticipantsWithPrivacy(t *testing.T) {
	t.Skip("Requires global db initialization - tested through handlers_test.go GetEventParticipants tests")
}

func TestApplyPrivacyFiltersEdgeCases(t *testing.T) {
	// Test 1: Non-participant viewer sees hidden organizer
	event := &Event{
		ID:                          1,
		UserID:                      100,
		Title:                       "Fully Private Event",
		CreatorName:                 "Organizer",
		UserEmail:                   "organizer@test.com",
		HideOrganizerUntilJoined:    true,
		HideParticipantsUntilJoined: true,
		IsParticipant:               false,
	}

	ApplyPrivacyFilters(event, 200, true, false) // viewer ID 200, verified, not admin
	assert.Contains(t, event.CreatorName, "Join to see")
	assert.Empty(t, event.UserEmail)

	// Test 2: Organizer always sees everything
	event2 := &Event{
		ID:                          1,
		UserID:                      100,
		Title:                       "Private Event",
		CreatorName:                 "Organizer",
		UserEmail:                   "organizer@test.com",
		HideOrganizerUntilJoined:    true,
		HideParticipantsUntilJoined: true,
	}

	ApplyPrivacyFilters(event2, 100, true, false) // organizer viewing own event
	assert.Equal(t, "Organizer", event2.CreatorName)
	assert.Equal(t, "organizer@test.com", event2.UserEmail)

	// Test 3: Admin always sees everything
	event3 := &Event{
		ID:                          1,
		UserID:                      100,
		Title:                       "Private Event",
		CreatorName:                 "Organizer",
		HideOrganizerUntilJoined:    true,
		HideParticipantsUntilJoined: true,
	}

	ApplyPrivacyFilters(event3, 200, true, true) // admin viewing
	assert.Equal(t, "Organizer", event3.CreatorName)

	// Test 4: Unverified user sees limited info
	event4 := &Event{
		ID:                       1,
		UserID:                   100,
		Title:                    "Event",
		CreatorName:              "Organizer",
		UserEmail:                "organizer@test.com",
		HideOrganizerUntilJoined: false,
		IsParticipant:            false,
	}

	ApplyPrivacyFilters(event4, 200, false, false) // unverified user
	assert.Empty(t, event4.UserEmail)
}

func TestCheckEventViewPermission(t *testing.T) {
	// Test 1: Unregistered user cannot view event that doesn't allow unregistered
	event := &Event{
		ID:                     1,
		AllowUnregisteredUsers: false,
	}
	reason := CheckEventViewPermission(event, 0, false, false) // userID=0 means unregistered
	assert.NotEmpty(t, reason)
	assert.Contains(t, reason, "registration")

	// Test 2: Registered unverified user cannot view event requiring verification
	event2 := &Event{
		ID:                     2,
		AllowUnregisteredUsers: false,
		RequireVerifiedToView:  true,
	}
	reason = CheckEventViewPermission(event2, 123, false, false) // userID=123, not verified
	assert.NotEmpty(t, reason)
	assert.Contains(t, reason, "verified email")

	// Test 3: Registered verified user can view
	reason = CheckEventViewPermission(event2, 123, true, false) // userID=123, verified
	assert.Empty(t, reason)

	// Test 4: Admin can always view
	reason = CheckEventViewPermission(event, 0, false, true) // unregistered but admin
	assert.Empty(t, reason)

	// Test 5: Event allowing unregistered users - anyone can view
	event3 := &Event{
		ID:                     3,
		AllowUnregisteredUsers: true,
		RequireVerifiedToView:  true, // This is ignored when allow_unregistered_users is true
	}
	reason = CheckEventViewPermission(event3, 0, false, false) // unregistered user
	assert.Empty(t, reason)

	// Test 6: Registered user can view event not requiring verification
	event4 := &Event{
		ID:                     4,
		AllowUnregisteredUsers: false,
		RequireVerifiedToView:  false,
	}
	reason = CheckEventViewPermission(event4, 123, false, false) // registered but unverified
	assert.Empty(t, reason)
}

func TestGetParticipantsWithPrivacyComprehensive(t *testing.T) {
	testDB := setupTestDB(t)
	defer cleanupTestDB(testDB)
	db = testDB

	// Create test users
	user1ID := createTestUser(t, testDB, "creator@example.com", "Creator", "password123", true)
	user2ID := createTestUser(t, testDB, "participant@example.com", "Participant", "password123", true)
	user3ID := createTestUser(t, testDB, "viewer@example.com", "Viewer", "password123", true)
	user4ID := createTestUser(t, testDB, "unverified@example.com", "Unverified", "password123", false)

	// Create test event with hide_participants_until_joined enabled
	eventID := createTestEvent(t, testDB, user1ID, "Test Event")
	testDB.Exec(`UPDATE events SET hide_participants_until_joined = 1 WHERE id = ?`, eventID)

	// Add participant
	testDB.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`, eventID, user2ID)

	t.Run("Creator can see participants", func(t *testing.T) {
		participants, err := GetParticipantsWithPrivacy(int(eventID), int(user1ID), true, false)
		assert.NoError(t, err)
		assert.NotEmpty(t, participants)
	})

	t.Run("Participant can see participants", func(t *testing.T) {
		participants, err := GetParticipantsWithPrivacy(int(eventID), int(user2ID), true, false)
		assert.NoError(t, err)
		assert.NotEmpty(t, participants)
	})

	t.Run("Non-participant viewer cannot see hidden participants", func(t *testing.T) {
		participants, err := GetParticipantsWithPrivacy(int(eventID), int(user3ID), true, false)
		assert.NoError(t, err)
		assert.Empty(t, participants)
	})

	t.Run("Admin can always see participants", func(t *testing.T) {
		participants, err := GetParticipantsWithPrivacy(int(eventID), int(user3ID), true, true)
		assert.NoError(t, err)
		assert.NotEmpty(t, participants)
	})

	t.Run("Unverified user sees participants without emails", func(t *testing.T) {
		// Create event without hiding participants
		eventID2 := createTestEvent(t, testDB, user1ID, "Public Event")
		testDB.Exec(`UPDATE events SET hide_participants_until_joined = 0 WHERE id = ?`, eventID2)
		testDB.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`, eventID2, user2ID)

		participants, err := GetParticipantsWithPrivacy(int(eventID2), int(user4ID), false, false)
		assert.NoError(t, err)
		assert.NotEmpty(t, participants)
		// Verify emails are hidden for unverified users
		for _, p := range participants {
			assert.Empty(t, p.Email)
		}
	})

	t.Run("Invalid event ID returns error", func(t *testing.T) {
		_, err := GetParticipantsWithPrivacy(99999, int(user1ID), true, false)
		assert.Error(t, err)
	})
}

func TestCheckEventJoinPermissionComprehensive(t *testing.T) {
	t.Run("Admin can join any event", func(t *testing.T) {
		event := Event{RequireVerifiedToJoin: true}
		result := CheckEventJoinPermission(&event, false, true)
		assert.Empty(t, result)
	})

	t.Run("Verified-only event blocks unverified users", func(t *testing.T) {
		event := Event{RequireVerifiedToJoin: true}
		result := CheckEventJoinPermission(&event, false, false)
		assert.Contains(t, result, "verified email")
	})

	t.Run("Verified user can join verified-only event", func(t *testing.T) {
		event := Event{RequireVerifiedToJoin: true}
		result := CheckEventJoinPermission(&event, true, false)
		assert.Empty(t, result)
	})

	t.Run("Anyone can join open event", func(t *testing.T) {
		event := Event{RequireVerifiedToJoin: false}
		result := CheckEventJoinPermission(&event, false, false)
		assert.Empty(t, result)
	})
}
