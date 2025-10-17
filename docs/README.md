# Veidly Documentation

This directory contains the Antora-based documentation for Veidly.

## Structure

```
docs/
├── antora.yml                    # Antora component configuration
└── modules/
    └── ROOT/
        ├── nav.adoc             # Navigation menu
        └── pages/
            ├── index.adoc       # Home page (README)
            ├── project-summary.adoc
            ├── changelog.adoc
            ├── guides/          # User guides
            │   ├── quickstart.adoc
            │   ├── new-features.adoc
            │   └── implementation.adoc
            ├── architecture/    # Architecture docs
            │   └── overview.adoc
            └── testing/         # Testing documentation
                ├── results.adoc
                ├── safety.adoc
                └── database-safety.adoc
```

## Building Documentation

### Option 1: Using Docker Compose (Recommended)

```bash
# Build the documentation
docker-compose -f docker-compose.docs.yml run antora

# Serve the documentation
docker-compose -f docker-compose.docs.yml up docs-server

# Access at: http://localhost:8000
```

### Option 2: Using Local Antora

```bash
# Install Antora (requires Node.js)
npm install -g @antora/cli @antora/site-generator

# Build documentation
antora antora-playbook.yml

# Serve with any static server
cd build/site
python3 -m http.server 8000
```

## Development

When adding new documentation:

1. Create `.adoc` files in appropriate `pages/` subdirectory
2. Add entry to `nav.adoc` for navigation
3. Use AsciiDoc format (similar to Markdown but more powerful)
4. Rebuild with Antora to see changes

## Converting Markdown to AsciiDoc

The files are currently in Markdown format (`.md`). To properly use Antora, they should be converted to AsciiDoc (`.adoc`).

### Manual Conversion Tips:
- Headers: `#` → `=`, `##` → `==`, `###` → `===`
- Code blocks: ` ```lang` → `[source,lang]` + `----`
- Links: `[text](url)` → `link:url[text]`
- Bold: `**text**` → `*text*`
- Italic: `*text*` → `_text_`

### Automated Conversion:
```bash
# Install pandoc
brew install pandoc  # macOS
apt-get install pandoc  # Linux

# Convert files
find docs/modules/ROOT/pages -name "*.adoc" -type f | while read file; do
  pandoc "$file" -f markdown -t asciidoc -o "$file"
done
```

## Documentation Standards

- Use clear, concise language
- Include code examples for features
- Keep navigation hierarchy flat (max 3 levels)
- Cross-reference related pages with `xref:`
- Add version information for API changes

## Resources

- [Antora Documentation](https://docs.antora.org)
- [AsciiDoc Syntax](https://docs.asciidoctor.org/asciidoc/latest/syntax-quick-reference/)
- [Antora UI Customization](https://docs.antora.org/antora-ui-default/)
