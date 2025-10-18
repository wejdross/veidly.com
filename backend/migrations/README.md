# Database Migrations

This directory contains one-time migration scripts to fix data issues.

## Completed Migrations

### 1. fix_html_entities.go
**Date**: 2025-10-18
**Purpose**: Decode HTML entities in event titles, descriptions, and creator names.

**Issue**: Event data stored in database contained HTML entities like `&#39;` (apostrophe), `&amp;` (ampersand), etc.

**What it fixes**:
- Event titles: `Mom&#39;s Book Club` → `Mom's Book Club`
- Descriptions: `Let&#39;s chat` → `Let's chat`
- Ampersands: `Museum &amp; Gallery` → `Museum & Gallery`

**Usage**:
```bash
cd backend/migrations
go run fix_html_entities.go
```

**Result**: Updated 11 events with HTML entities.

---

### 2. fix_slugs.go
**Date**: 2025-10-18
**Purpose**: Regenerate event slugs to remove HTML entity codes.

**Issue**: Event slugs contained HTML entity numbers like `mom-39-s-book-club` instead of `mom-s-book-club`.

**What it fixes**:
- Slugs with `39` (apostrophe): `mom-39-s-book-club` → `mom-s-book-club`
- Slugs with `amp`: `vineyard-visit-amp-lunch` → `vineyard-visit-lunch`

**Usage**:
```bash
cd backend/migrations
go run fix_slugs.go
```

**Result**: Updated 7 event slugs.

**Note**: This migration preserves the random suffix at the end of each slug to maintain URL uniqueness.

---

## Running Migrations

These migrations have already been run on the development database. For production:

1. **Backup the database first**:
   ```bash
   sqlite3 /var/lib/veidly/veidly.db ".backup '/home/veidly/backups/pre_migration_$(date +%Y%m%d_%H%M%S).db'"
   ```

2. **Run the migration**:
   ```bash
   cd /path/to/backend/migrations
   go run fix_html_entities.go
   go run fix_slugs.go
   ```

3. **Verify the changes**:
   ```bash
   sqlite3 /var/lib/veidly/veidly.db "SELECT id, title, slug FROM events WHERE id = 7;"
   ```

## Prevention

The root cause has been fixed in the codebase:
- **Backend**: `generateSlug()` function now decodes HTML entities before processing (utils.go)
- **Frontend**: Event titles are now rendered as plain text instead of using `dangerouslySetInnerHTML`

New events created after these fixes will not have HTML entity issues.
