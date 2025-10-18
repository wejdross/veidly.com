// Centralized API configuration
// VITE_API_URL should point to the base URL without /api (e.g., http://localhost:8080)
// In production, if VITE_API_URL is empty, use relative URLs (same domain as frontend)
const BASE_URL = import.meta.env.VITE_API_URL || (
  // Use relative URL in production (when served from same domain)
  import.meta.env.MODE === 'production' ? '' : 'http://localhost:8080'
)
export const API_BASE_URL = `${BASE_URL}/api`  // For /api/* endpoints
export const API_BASE_URL_ROOT = BASE_URL || ''  // For root-level endpoints
