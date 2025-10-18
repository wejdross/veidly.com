package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		expectedPrefix string
	}{
		{
			name:        "Simple title",
			title:       "Hello World",
			expectedPrefix: "hello-world-",
		},
		{
			name:        "Title with special characters",
			title:       "Hello, World!",
			expectedPrefix: "hello-world-",
		},
		{
			name:        "Title with multiple spaces",
			title:       "Hello    World",
			expectedPrefix: "hello-world-",
		},
		{
			name:        "Title with unicode",
			title:       "Café in Paris",
			expectedPrefix: "caf-in-paris-",
		},
		{
			name:        "Title with Polish characters",
			title:       "Spotkanie w Łodzi",
			expectedPrefix: "spotkanie-w-odzi-",
		},
		{
			name:        "Title starting and ending with spaces",
			title:       "  Hello World  ",
			expectedPrefix: "hello-world-",
		},
		{
			name:        "Title with numbers",
			title:       "Event 2025",
			expectedPrefix: "event-2025-",
		},
		{
			name:        "Title with hyphens",
			title:       "Pre-event Meeting",
			expectedPrefix: "pre-event-meeting-",
		},
		{
			name:        "Title with HTML entities",
			title:       "Mom&#39;s Book Club",
			expectedPrefix: "mom-s-book-club-",
		},
		{
			name:        "Title with various HTML entities",
			title:       "Coffee &amp; Chat &#8211; Let&#39;s Meet!",
			expectedPrefix: "coffee-chat-let-s-meet-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSlug(tt.title)
			assert.True(t, strings.HasPrefix(result, tt.expectedPrefix), "Expected slug to start with %s, got %s", tt.expectedPrefix, result)
			// Verify it has a random suffix (8 characters)
			assert.True(t, len(result) > len(tt.expectedPrefix), "Slug should have random suffix")
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	// Test length
	length := 16
	result := generateRandomString(length)
	assert.Equal(t, length, len(result))

	// Test different lengths
	for _, l := range []int{8, 16, 32, 64} {
		result := generateRandomString(l)
		assert.Equal(t, l, len(result))
	}

	// Test uniqueness (generate 100 strings and ensure they're all different)
	generated := make(map[string]bool)
	for i := 0; i < 100; i++ {
		result := generateRandomString(32)
		assert.False(t, generated[result], "Generated duplicate random string")
		generated[result] = true
	}

	// Test that it only contains valid characters (alphanumeric)
	result = generateRandomString(100)
	for _, char := range result {
		assert.True(t, (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9'),
			"Invalid character in random string: %c", char)
	}
}

func TestGenerateUniqueSlug(t *testing.T) {
	t.Skip("Requires global db initialization - tested through handlers_test.go event creation tests")
}

func TestSlugSanitization(t *testing.T) {
	// Test that generateSlug sanitizes dangerous characters
	maliciousTitle := "Test'; DROP TABLE events; --"
	slug := generateSlug(maliciousTitle)

	// Slug should be sanitized
	assert.NotContains(t, slug, "'")
	assert.NotContains(t, slug, ";")
	assert.NotContains(t, slug, "--")
	assert.True(t, strings.HasPrefix(slug, "test-drop-table-events"))
}
