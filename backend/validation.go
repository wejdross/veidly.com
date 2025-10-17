package main

import (
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Validation errors
var (
	ErrTitleTooLong       = errors.New("title too long (max 200 characters)")
	ErrTitleTooShort      = errors.New("title too short (min 3 characters)")
	ErrDescriptionTooLong = errors.New("description too long (max 5000 characters)")
	ErrDescriptionTooShort = errors.New("description too short (min 10 characters)")
	ErrInvalidLatitude    = errors.New("invalid latitude (must be between -90 and 90)")
	ErrInvalidLongitude   = errors.New("invalid longitude (must be between -180 and 180)")
	ErrInvalidParticipants = errors.New("max_participants must be positive or zero")
	ErrInvalidAgeRange    = errors.New("age_min must be less than or equal to age_max")
	ErrInvalidAgeValues   = errors.New("age values must be between 0 and 150")
	ErrEventInPast        = errors.New("event cannot start in the past (more than 1 hour ago)")
	ErrEndBeforeStart     = errors.New("end time must be after start time")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong    = errors.New("password must be less than 100 characters")
	ErrNameTooShort       = errors.New("name must be at least 2 characters")
	ErrNameTooLong        = errors.New("name too long (max 100 characters)")
	ErrInvalidContact     = errors.New("contact method too short (min 3 characters)")
)

// Email regex for basic validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEvent validates event data before creation/update
func ValidateEvent(event *Event, startTime, endTime *time.Time) error {
	// Title validation
	if utf8.RuneCountInString(event.Title) < 3 {
		return ErrTitleTooShort
	}
	if utf8.RuneCountInString(event.Title) > 200 {
		return ErrTitleTooLong
	}

	// Description validation
	if utf8.RuneCountInString(event.Description) < 10 {
		return ErrDescriptionTooShort
	}
	if utf8.RuneCountInString(event.Description) > 5000 {
		return ErrDescriptionTooLong
	}

	// Coordinate validation
	if event.Latitude < -90 || event.Latitude > 90 {
		return ErrInvalidLatitude
	}
	if event.Longitude < -180 || event.Longitude > 180 {
		return ErrInvalidLongitude
	}

	// Participants validation
	if event.MaxParticipants < 0 {
		return ErrInvalidParticipants
	}

	// Age validation
	if event.AgeMin < 0 || event.AgeMin > 150 || event.AgeMax < 0 || event.AgeMax > 150 {
		return ErrInvalidAgeValues
	}
	if event.AgeMin > event.AgeMax {
		return ErrInvalidAgeRange
	}

	// Time validation
	if startTime != nil {
		// Allow events starting up to 1 hour in the past (for timezone/clock differences)
		if startTime.Before(time.Now().Add(-1 * time.Hour)) {
			return ErrEventInPast
		}

		// Validate end time if provided
		if endTime != nil && endTime.Before(*startTime) {
			return ErrEndBeforeStart
		}
	}

	// Category validation
	validCategory := false
	for _, cat := range Categories {
		if event.Category == cat {
			validCategory = true
			break
		}
	}
	if !validCategory {
		return fmt.Errorf("invalid category: %s", event.Category)
	}

	// Gender restriction validation
	validGender := []string{"any", "male", "female", "non-binary"}
	validGenderRestriction := false
	for _, g := range validGender {
		if event.GenderRestriction == g {
			validGenderRestriction = true
			break
		}
	}
	if !validGenderRestriction {
		return fmt.Errorf("invalid gender_restriction: %s", event.GenderRestriction)
	}

	// Creator name validation
	if utf8.RuneCountInString(event.CreatorName) < 2 {
		return ErrNameTooShort
	}
	if utf8.RuneCountInString(event.CreatorName) > 100 {
		return ErrNameTooLong
	}

	// Sanitize HTML to prevent XSS
	event.Title = html.EscapeString(event.Title)
	event.Description = html.EscapeString(event.Description)
	event.CreatorName = html.EscapeString(event.CreatorName)

	return nil
}

// ValidateUser validates user data during registration
func ValidateUser(user *User) error {
	// Email validation
	email := strings.TrimSpace(user.Email)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	user.Email = strings.ToLower(email)

	// Password validation
	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}
	if len(user.Password) > 100 {
		return ErrPasswordTooLong
	}

	// Name validation
	if utf8.RuneCountInString(user.Name) < 2 {
		return ErrNameTooShort
	}
	if utf8.RuneCountInString(user.Name) > 100 {
		return ErrNameTooLong
	}

	// Sanitize fields
	user.Name = html.EscapeString(user.Name)
	if user.Bio != "" {
		if utf8.RuneCountInString(user.Bio) > 1000 {
			return errors.New("bio too long (max 1000 characters)")
		}
		user.Bio = html.EscapeString(user.Bio)
	}

	return nil
}

// ValidateProfileUpdate validates profile update data
func ValidateProfileUpdate(req *ProfileUpdateRequest) error {
	// Name validation
	if req.Name != "" {
		if utf8.RuneCountInString(req.Name) < 2 {
			return ErrNameTooShort
		}
		if utf8.RuneCountInString(req.Name) > 100 {
			return ErrNameTooLong
		}
		req.Name = html.EscapeString(req.Name)
	}

	// Bio validation
	if req.Bio != "" {
		if utf8.RuneCountInString(req.Bio) > 1000 {
			return errors.New("bio too long (max 1000 characters)")
		}
		req.Bio = html.EscapeString(req.Bio)
	}

	// Default contact method validation (basic)
	if req.Threema != "" {
		req.Threema = html.EscapeString(req.Threema)
	}

	return nil
}
