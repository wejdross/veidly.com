package main

import (
	"database/sql"
	"fmt"
	"html"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "./veidly.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Find all events with HTML entities in titles or descriptions
	rows, err := db.Query(`
		SELECT id, title, description, creator_name
		FROM events
		WHERE title LIKE '%&#%'
		   OR title LIKE '%&amp;%'
		   OR title LIKE '%&quot;%'
		   OR description LIKE '%&#%'
		   OR description LIKE '%&amp;%'
		   OR description LIKE '%&quot;%'
		   OR creator_name LIKE '%&#%'
		   OR creator_name LIKE '%&amp;%'
		   OR creator_name LIKE '%&quot;%'
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var updates []struct {
		id              int
		oldTitle        string
		newTitle        string
		oldDescription  string
		newDescription  string
		oldCreatorName  string
		newCreatorName  string
		hasChanges      bool
	}

	// Collect all updates
	for rows.Next() {
		var id int
		var title, description, creatorName string
		if err := rows.Scan(&id, &title, &description, &creatorName); err != nil {
			log.Fatal(err)
		}

		newTitle := html.UnescapeString(title)
		newDescription := html.UnescapeString(description)
		newCreatorName := html.UnescapeString(creatorName)

		hasChanges := (title != newTitle) || (description != newDescription) || (creatorName != newCreatorName)

		if hasChanges {
			updates = append(updates, struct {
				id              int
				oldTitle        string
				newTitle        string
				oldDescription  string
				newDescription  string
				oldCreatorName  string
				newCreatorName  string
				hasChanges      bool
			}{
				id:             id,
				oldTitle:       title,
				newTitle:       newTitle,
				oldDescription: description,
				newDescription: newDescription,
				oldCreatorName: creatorName,
				newCreatorName: newCreatorName,
				hasChanges:     hasChanges,
			})
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Display what will be updated
	fmt.Printf("Found %d events with HTML entities:\n\n", len(updates))
	for _, u := range updates {
		fmt.Printf("Event ID %d:\n", u.id)
		if u.oldTitle != u.newTitle {
			fmt.Printf("  Title:\n")
			fmt.Printf("    Old: %s\n", u.oldTitle)
			fmt.Printf("    New: %s\n", u.newTitle)
		}
		if u.oldDescription != u.newDescription {
			fmt.Printf("  Description:\n")
			fmt.Printf("    Old: %.100s...\n", u.oldDescription)
			fmt.Printf("    New: %.100s...\n", u.newDescription)
		}
		if u.oldCreatorName != u.newCreatorName {
			fmt.Printf("  Creator Name:\n")
			fmt.Printf("    Old: %s\n", u.oldCreatorName)
			fmt.Printf("    New: %s\n", u.newCreatorName)
		}
		fmt.Println()
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
		_, err := db.Exec(
			"UPDATE events SET title = ?, description = ?, creator_name = ? WHERE id = ?",
			u.newTitle, u.newDescription, u.newCreatorName, u.id,
		)
		if err != nil {
			log.Printf("Failed to update event %d: %v", u.id, err)
			continue
		}
		updated++
	}

	fmt.Printf("\n✅ Successfully updated %d events!\n", updated)
}
