package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
)

// generateSlug creates a URL-friendly slug from a title
func generateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit length to 50 characters
	if len(slug) > 50 {
		slug = slug[:50]
		slug = strings.TrimRight(slug, "-")
	}

	// Add random suffix to ensure uniqueness
	suffix := generateRandomString(8)
	slug = slug + "-" + suffix

	return slug
}

// generateRandomString creates a random string of specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length/2+1)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// generateUniqueSlug generates a unique slug with retry logic
func generateUniqueSlug(title string) (string, error) {
	for i := 0; i < 5; i++ {
		slug := generateSlug(title)
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM events WHERE slug = ?", slug).Scan(&count)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
		// If collision, try again with different random suffix
	}
	return "", errors.New("failed to generate unique slug after 5 attempts")
}
