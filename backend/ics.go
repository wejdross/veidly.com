package main

import (
	"fmt"
	"strings"
	"time"
)

// GenerateICS creates an ICS (iCalendar) file content for an event
func GenerateICS(event *Event) string {
	// Parse start time
	startTime, err := time.Parse(time.RFC3339, event.StartTime)
	if err != nil {
		// Try alternative format
		startTime, _ = time.Parse("2006-01-02T15:04:05Z07:00", event.StartTime)
	}

	// Parse end time if available, otherwise default to 2 hours after start
	var endTime time.Time
	if event.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, event.EndTime)
		if err != nil {
			endTime, _ = time.Parse("2006-01-02T15:04:05Z07:00", event.EndTime)
		}
	} else {
		endTime = startTime.Add(2 * time.Hour)
	}

	// Format times for ICS (YYYYMMDDTHHMMSSZ)
	startICS := startTime.UTC().Format("20060102T150405Z")
	endICS := endTime.UTC().Format("20060102T150405Z")
	nowICS := time.Now().UTC().Format("20060102T150405Z")

	// Generate unique UID
	uid := fmt.Sprintf("event-%d@veidly.com", event.ID)

	// Clean and escape text for ICS format
	title := escapeICS(event.Title)
	description := escapeICS(event.Description)
	location := fmt.Sprintf("%.6f,%.6f", event.Latitude, event.Longitude)
	organizer := escapeICS(event.CreatorName)

	// Build ICS content
	ics := strings.Builder{}
	ics.WriteString("BEGIN:VCALENDAR\r\n")
	ics.WriteString("VERSION:2.0\r\n")
	ics.WriteString("PRODID:-//Veidly//Event Calendar//EN\r\n")
	ics.WriteString("CALSCALE:GREGORIAN\r\n")
	ics.WriteString("METHOD:PUBLISH\r\n")
	ics.WriteString("BEGIN:VEVENT\r\n")
	ics.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
	ics.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", nowICS))
	ics.WriteString(fmt.Sprintf("DTSTART:%s\r\n", startICS))
	ics.WriteString(fmt.Sprintf("DTEND:%s\r\n", endICS))
	ics.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", title))
	ics.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", description))
	ics.WriteString(fmt.Sprintf("LOCATION:%s\r\n", location))
	ics.WriteString(fmt.Sprintf("ORGANIZER;CN=%s:MAILTO:noreply@veidly.com\r\n", organizer))
	ics.WriteString("STATUS:CONFIRMED\r\n")
	ics.WriteString("SEQUENCE:0\r\n")

	// Add categories based on event category
	if event.Category != "" {
		categoryName := CategoryNames[event.Category]
		if categoryName != "" {
			ics.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", escapeICS(categoryName)))
		}
	}

	// Add URL if slug is available
	if event.Slug != "" {
		ics.WriteString(fmt.Sprintf("URL:https://veidly.com/event/%s\r\n", event.Slug))
	}

	ics.WriteString("END:VEVENT\r\n")
	ics.WriteString("END:VCALENDAR\r\n")

	return ics.String()
}

// escapeICS escapes special characters for ICS format
func escapeICS(text string) string {
	// Replace special characters according to RFC 5545
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ";", "\\;")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\r", "")
	return text
}
