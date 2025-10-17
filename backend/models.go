package main

import "time"

type User struct {
	ID             int       `json:"id"`
	Email          string    `json:"email" binding:"required,email"`
	Password       string    `json:"-" binding:"required,min=8"` // Never expose password in JSON responses
	Name           string    `json:"name" binding:"required"`
	Bio            string    `json:"bio"`
	Threema        string    `json:"threema"`
	Languages      string    `json:"languages"` // Comma-separated language codes (e.g., "en,de,fr")
	IsAdmin        bool      `json:"is_admin"`
	IsBlocked      bool      `json:"is_blocked"`
	EmailVerified  bool      `json:"email_verified"`
	CreatedAt      time.Time `json:"created_at"`
}

type ProfileUpdateRequest struct {
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	Threema   string `json:"threema"`
	Languages string `json:"languages"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type EmailVerificationToken struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type PasswordResetToken struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Used      bool      `json:"used"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type Event struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	Title             string    `json:"title" binding:"required"`
	Description       string    `json:"description" binding:"required"`
	Category          string    `json:"category" binding:"required"`
	Latitude          float64   `json:"latitude" binding:"required"`
	Longitude         float64   `json:"longitude" binding:"required"`
	StartTime         string    `json:"start_time" binding:"required"`
	EndTime           string    `json:"end_time"`
	CreatorName       string    `json:"creator_name" binding:"required"`
	MaxParticipants   int       `json:"max_participants"`
	GenderRestriction string    `json:"gender_restriction"`
	AgeMin            int       `json:"age_min"`
	AgeMax            int       `json:"age_max"`
	SmokingAllowed    bool      `json:"smoking_allowed"`
	AlcoholAllowed    bool      `json:"alcohol_allowed"`
	EventLanguages    string    `json:"event_languages"` // Comma-separated language codes for the event
	Slug              string    `json:"slug"`
	CreatedAt         time.Time `json:"created_at"`

	// Privacy controls
	HideOrganizerUntilJoined    bool `json:"hide_organizer_until_joined"`
	HideParticipantsUntilJoined bool `json:"hide_participants_until_joined"`
	RequireVerifiedToJoin       bool `json:"require_verified_to_join"`
	RequireVerifiedToView       bool `json:"require_verified_to_view"`

	// Joined data
	UserEmail        string `json:"user_email,omitempty"`
	CreatorLanguages string `json:"creator_languages,omitempty"`
	ParticipantCount int    `json:"participant_count"`
	Participants     []User `json:"participants,omitempty"`
	IsParticipant    bool   `json:"is_participant,omitempty"` // Whether current user is a participant
}

type EventParticipant struct {
	ID       int       `json:"id"`
	EventID  int       `json:"event_id"`
	UserID   int       `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type Place struct {
	DisplayName string  `json:"display_name"`
	Lat         string  `json:"lat"`
	Lon         string  `json:"lon"`
	Type        string  `json:"type"`
	Importance  float64 `json:"importance"`
}

var Categories = []string{
	"social_drinks",       // Social & Drinks 🍻
	"sports_fitness",      // Sports & Fitness 🏃
	"food_dining",         // Food & Dining 🍕
	"business_networking", // Business & Networking 💼
	"gaming_hobbies",      // Gaming & Hobbies 🎮
	"learning_skills",     // Learning & Skills 📚
	"adventure_travel",    // Adventure & Travel ✈️
	"parents_kids",        // Parents & Kids 👶
}

var CategoryNames = map[string]string{
	"social_drinks":       "Social & Drinks 🍻",
	"sports_fitness":      "Sports & Fitness 🏃",
	"food_dining":         "Food & Dining 🍕",
	"business_networking": "Business & Networking 💼",
	"gaming_hobbies":      "Gaming & Hobbies 🎮",
	"learning_skills":     "Learning & Skills 📚",
	"adventure_travel":    "Adventure & Travel ✈️",
	"parents_kids":        "Parents & Kids 👶",
}

// UserBlock represents a user blocking relationship
type UserBlock struct {
	ID        int       `json:"id"`
	BlockerID int       `json:"blocker_id"`
	BlockedID int       `json:"blocked_id"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// BlockUserRequest represents the request to block a user
type BlockUserRequest struct {
	Reason string `json:"reason,omitempty"`
}

// EventComment represents a comment on an event
type EventComment struct {
	ID        int       `json:"id"`
	EventID   int       `json:"event_id"`
	UserID    int       `json:"user_id"`
	Comment   string    `json:"comment" binding:"required"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	IsDeleted bool      `json:"is_deleted"`
	IsOwn     bool      `json:"is_own"`
}

// CreateCommentRequest represents the request to create a comment
type CreateCommentRequest struct {
	Comment string `json:"comment" binding:"required,min=1,max=1000"`
}

// UpdateCommentRequest represents the request to update a comment
type UpdateCommentRequest struct {
	Comment string `json:"comment" binding:"required,min=1,max=1000"`
}

// EventReport represents a report on an event
type EventReport struct {
	ID          int       `json:"id"`
	EventID     int       `json:"event_id"`
	ReporterID  int       `json:"reporter_id"`
	Reason      string    `json:"reason" binding:"required"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateReportRequest represents the request to create a report
type CreateReportRequest struct {
	Reason      string `json:"reason" binding:"required"`
	Description string `json:"description,omitempty"`
}

// CommentReport represents a report on a comment
type CommentReport struct {
	ID          int       `json:"id"`
	CommentID   int       `json:"comment_id"`
	ReporterID  int       `json:"reporter_id"`
	Reason      string    `json:"reason" binding:"required"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
