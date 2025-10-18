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
