package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateProfileUpdate(t *testing.T) {
	tests := []struct {
		name        string
		req         ProfileUpdateRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid profile update",
			req: ProfileUpdateRequest{
				Name:      "John Doe",
				Bio:       "Software developer interested in hiking",
				Languages: "English,Spanish",
			},
			wantErr: false,
		},
		{
			name: "Name too short",
			req: ProfileUpdateRequest{
				Name: "J",
				Bio:  "Some bio",
			},
			wantErr:     true,
			errContains: "at least 2 characters",
		},
		{
			name: "Name too long",
			req: ProfileUpdateRequest{
				Name: "This is an extremely long display name that exceeds the one hundred character limit and should fail validation",
				Bio:  "Some bio",
			},
			wantErr:     true,
			errContains: "name too long",
		},
		{
			name: "Bio too long",
			req: ProfileUpdateRequest{
				Name: "Valid Name",
				Bio:  string(make([]byte, 1001)), // 1001 characters
			},
			wantErr:     true,
			errContains: "bio too long",
		},
		{
			name: "Minimal valid profile",
			req: ProfileUpdateRequest{
				Name: "Jo",
				Bio:  "",
			},
			wantErr: false,
		},
		{
			name: "Valid with phone",
			req: ProfileUpdateRequest{
				Name:  "John",
			},
			wantErr: false,
		},
		{
			name: "Valid with bio",
			req: ProfileUpdateRequest{
				Name:    "John",
				Bio:     "Test bio",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfileUpdate(&tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEventComprehensive(t *testing.T) {
	startTime, _ := time.Parse(time.DateTime, "2025-12-01 10:00:00")
	endTime, _ := time.Parse(time.DateTime, "2025-12-01 12:00:00")
	invalidEndTime, _ := time.Parse(time.DateTime, "2025-12-01 09:00:00") // before start

	tests := []struct {
		name        string
		event       Event
		startTime   *time.Time
		endTime     *time.Time
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid event",
			event: Event{
				Title:              "Test Event",
				Description:        "A great event description",
				Category:           "social_drinks",
				Latitude:           52.2297,
				Longitude:          21.0122,
				GenderRestriction:  "any",
				CreatorName:        "John Doe",
			},
			startTime: &startTime,
			endTime:   &endTime,
			wantErr:   false,
		},
		{
			name: "Title too short",
			event: Event{
				Title:       "Hi",
				Description: "Valid description",
				Category:    "sports_fitness",
				Latitude:    52.2297,
				Longitude:   21.0122,
			},
			startTime:   &startTime,
			endTime:     &endTime,
			wantErr:     true,
			errContains: "title",
		},
		{
			name: "Description too short",
			event: Event{
				Title:       "Valid Title",
				Description: "Short",
				Category:    "food_dining",
				Latitude:    52.2297,
				Longitude:   21.0122,
			},
			startTime:   &startTime,
			endTime:     &endTime,
			wantErr:     true,
			errContains: "description",
		},
		{
			name: "Invalid category",
			event: Event{
				Title:       "Valid Title",
				Description: "Valid description here",
				Category:    "invalid_category",
				Latitude:    52.2297,
				Longitude:   21.0122,
			},
			startTime:   &startTime,
			endTime:     &endTime,
			wantErr:     true,
			errContains: "category",
		},
		{
			name: "Latitude out of range",
			event: Event{
				Title:       "Valid Title",
				Description: "Valid description",
				Category:    "social_drinks",
				Latitude:    91.0,
				Longitude:   21.0122,
			},
			startTime:   &startTime,
			endTime:     &endTime,
			wantErr:     true,
			errContains: "latitude",
		},
		{
			name: "Longitude out of range",
			event: Event{
				Title:       "Valid Title",
				Description: "Valid description",
				Category:    "social_drinks",
				Latitude:    52.2297,
				Longitude:   -181.0,
			},
			startTime:   &startTime,
			endTime:     &endTime,
			wantErr:     true,
			errContains: "longitude",
		},
		{
			name: "End time before start time",
			event: Event{
				Title:       "Valid Title",
				Description: "Valid description",
				Category:    "social_drinks",
				Latitude:    52.2297,
				Longitude:   21.0122,
			},
			startTime:   &startTime,
			endTime:     &invalidEndTime,
			wantErr:     true,
			errContains: "end time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEvent(&tt.event, tt.startTime, tt.endTime)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUserComprehensive(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid user",
			user: User{
				Email:    "test@example.com",
				Password: "SecurePass123",
				Name:     "John Doe",
			},
			wantErr: false,
		},
		{
			name: "Invalid email format",
			user: User{
				Email:    "not-an-email",
				Password: "SecurePass123",
				Name:     "John Doe",
			},
			wantErr:     true,
			errContains: "email",
		},
		{
			name: "Password too short",
			user: User{
				Email:    "test@example.com",
				Password: "short",
				Name:     "John Doe",
			},
			wantErr:     true,
			errContains: "password",
		},
		{
			name: "Display name too short",
			user: User{
				Email:    "test@example.com",
				Password: "SecurePass123",
				Name:     "J",
			},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "Display name too long",
			user: User{
				Email:    "test@example.com",
				Password: "SecurePass123",
				Name:     "This is an extremely long display name that definitely exceeds the one hundred character limit for names",
			},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "Minimal valid password length",
			user: User{
				Email:    "test@example.com",
				Password: "12345678",
				Name:     "John",
			},
			wantErr: false,
		},
		{
			name: "Minimal valid display name",
			user: User{
				Email:    "test@example.com",
				Password: "SecurePass123",
				Name:     "Jo",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUser(&tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
