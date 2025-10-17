import DOMPurify from 'dompurify'

/**
 * Sanitizes HTML content, allowing only safe tags
 * Use this for user-generated content that may contain HTML formatting
 */
export function sanitizeHTML(dirty: string): string {
  return DOMPurify.sanitize(dirty, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'br', 'p', 'a'],
    ALLOWED_ATTR: ['href', 'target', 'rel'],
    ALLOW_DATA_ATTR: false,
  })
}

/**
 * Escapes HTML in plain text content
 * Use this for user-generated text that should be displayed as-is without HTML
 */
export function sanitizeText(text: string): string {
  // Create a temporary div element to escape HTML
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

/**
 * Sanitizes URL to prevent javascript: and data: protocols
 * Use this before setting href attributes
 */
export function sanitizeURL(url: string): string {
  const urlLower = url.toLowerCase().trim()

  // Block dangerous protocols
  if (
    urlLower.startsWith('javascript:') ||
    urlLower.startsWith('data:') ||
    urlLower.startsWith('vbscript:') ||
    urlLower.startsWith('file:')
  ) {
    return '#'
  }

  return url
}
