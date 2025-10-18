// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"html"
	"log"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// regenerateSlugFromTitle regenerates a slug from title without the random suffix
func regenerateSlugFromTitle(title string, existingSuffix string) string {
	// Decode HTML entities first (e.g., &#39; -> ')
	slug := html.UnescapeString(title)

	// Convert to lowercase
	slug = strings.ToLower(slug)

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

	// Keep the existing random suffix to maintain URL uniqueness
	slug = slug + "-" + existingSuffix

	return slug
}

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "./veidly.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Find all events with HTML entities in slugs
	rows, err := db.Query(`
		SELECT id, title, slug
		FROM events
		WHERE slug LIKE '%39%'
		   OR slug LIKE '%amp%'
		   OR slug LIKE '%quot%'
		   OR slug LIKE '%lt%'
		   OR slug LIKE '%gt%'
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var updates []struct {
		id      int
		oldSlug string
		newSlug string
	}

	// Collect all updates
	for rows.Next() {
		var id int
		var title, oldSlug string
		if err := rows.Scan(&id, &title, &oldSlug); err != nil {
			log.Fatal(err)
		}

		// Extract the random suffix from existing slug (last 8 characters after last hyphen)
		parts := strings.Split(oldSlug, "-")
		if len(parts) == 0 {
			log.Printf("Skipping event %d: invalid slug format", id)
			continue
		}
		suffix := parts[len(parts)-1]

		// Generate new slug with same suffix
		newSlug := regenerateSlugFromTitle(title, suffix)

		if oldSlug != newSlug {
			updates = append(updates, struct {
				id      int
				oldSlug string
				newSlug string
			}{id, oldSlug, newSlug})
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Display what will be updated
	fmt.Printf("Found %d events with HTML entities in slugs:\n\n", len(updates))
	for _, u := range updates {
		fmt.Printf("Event ID %d:\n", u.id)
		fmt.Printf("  Old: %s\n", u.oldSlug)
		fmt.Printf("  New: %s\n\n", u.newSlug)
	}

	if len(updates) == 0 {
		fmt.Println("No events need updating!")
		return
	}

	// Ask for confirmation
	fmt.Print("Proceed with updates? (yes/no): ")
	var response string
	fmt.Scanln(&response)

	if response != "yes" {
		fmt.Println("Aborted.")
		return
	}

	// Perform updates
	updated := 0
	for _, u := range updates {
		_, err := db.Exec("UPDATE events SET slug = ? WHERE id = ?", u.newSlug, u.id)
		if err != nil {
			log.Printf("Failed to update event %d: %v", u.id, err)
			continue
		}
		updated++
	}

	fmt.Printf("\n✅ Successfully updated %d slugs!\n", updated)
}
