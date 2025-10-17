# Veidly Logo & Brand Assets

## Logo Files

### 1. **logo.svg** - Full Icon (200x200)
- **Use for:** App icons, social media profiles, marketing materials
- High detail version with two people meeting inside a map pin
- Includes subtle highlights and connection spark

### 2. **favicon.svg** - Simplified Icon (64x64)
- **Use for:** Browser favicon, app shortcuts, small UI elements
- Optimized for small sizes (16px - 64px)
- Simplified design: two dots connected inside a map pin

### 3. **logo-with-text.svg** - Logo + Wordmark (400x120)
- **Use for:** Headers, landing pages, email signatures
- Includes icon + "Veidly" text + tagline
- Best for horizontal layouts

## Design Concept

### Symbol Meaning
The logo combines three key elements:

1. **Map Pin** 🗺️
   - Represents location-based meetups
   - Shows "real world" connections, not virtual
   - Points to specific places where people meet

2. **Two People** 👥
   - Human connection at the core
   - Face-to-face interaction
   - Community and friendship

3. **Connection Spark** ✨
   - The moment people connect
   - Authentic human interaction
   - Breaking through digital noise

### Brand Colors

```css
Primary Purple:   #667eea
Secondary Purple: #764ba2
Gradient:         linear-gradient(135deg, #667eea 0%, #764ba2 100%)
```

The purple gradient represents:
- **Trust** and **authenticity**
- **Creativity** and **uniqueness**
- **Community** and **connection**

### Typography

- **Font Family:** System UI fonts (system-ui, -apple-system, 'Segoe UI', sans-serif)
- **Wordmark Weight:** 800 (Extra Bold)
- **Tagline Weight:** 600 (Semi Bold)

## Usage Guidelines

### ✅ Do's
- Use on white or light backgrounds for best contrast
- Maintain aspect ratio when scaling
- Provide adequate spacing around the logo (minimum 10px)
- Use SVG format for crisp rendering at all sizes

### ❌ Don'ts
- Don't rotate or skew the logo
- Don't change the gradient colors
- Don't add drop shadows or effects
- Don't place on busy backgrounds that reduce readability

## File Formats

All logos are in **SVG** format for:
- ✅ Infinite scalability without quality loss
- ✅ Small file sizes
- ✅ Sharp rendering on all screens
- ✅ Easy color modifications via CSS

## Implementation

The favicon is automatically loaded in `index.html`:

```html
<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
<link rel="apple-touch-icon" href="/logo.svg" />
<meta name="theme-color" content="#667eea" />
```

## Tagline

**"Real People, Real Connections"**

This tagline emphasizes:
- Authenticity in an age of AI-generated content
- Human-to-human interaction
- Genuine relationships vs. fake social media engagement
- Real-world meetups vs. virtual interactions
