-- Migration: Add privacy features to events table
-- Created: 2025-10-15
-- Description: Adds privacy control columns to allow event creators to control information visibility

-- Add privacy control columns to events table
ALTER TABLE events ADD COLUMN hide_organizer_until_joined BOOLEAN DEFAULT 0;
ALTER TABLE events ADD COLUMN hide_participants_until_joined BOOLEAN DEFAULT 1;
ALTER TABLE events ADD COLUMN require_verified_to_join BOOLEAN DEFAULT 0;
ALTER TABLE events ADD COLUMN require_verified_to_view BOOLEAN DEFAULT 0;

-- Create index for verified user filtering
CREATE INDEX IF NOT EXISTS idx_events_require_verified ON events(require_verified_to_view);

-- Update existing events to have default privacy settings
-- Default: participants hidden until joined, organizer visible, no verification required
UPDATE events
SET
  hide_organizer_until_joined = 0,
  hide_participants_until_joined = 1,
  require_verified_to_join = 0,
  require_verified_to_view = 0
WHERE hide_organizer_until_joined IS NULL;
