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
	// Test 1: Unverified user cannot view event requiring verification
	event := &Event{
		ID:                    1,
		RequireVerifiedToView: true,
	}
	reason := CheckEventViewPermission(event, false, false)
	assert.NotEmpty(t, reason)
	assert.Contains(t, reason, "verified email")

	// Test 2: Verified user can view
	reason = CheckEventViewPermission(event, true, false)
	assert.Empty(t, reason)

	// Test 3: Admin can always view
	reason = CheckEventViewPermission(event, false, true)
	assert.Empty(t, reason)

	// Test 4: Public event (anyone can view)
	event2 := &Event{
		ID:                    2,
		RequireVerifiedToView: false,
	}
	reason = CheckEventViewPermission(event2, false, false)
	assert.Empty(t, reason)
}
