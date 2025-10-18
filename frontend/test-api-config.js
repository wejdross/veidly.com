#!/usr/bin/env node
/**
 * Test script to verify API configuration in different modes
 * Run: node test-api-config.js
 */

console.log('🧪 Testing API Configuration\n');

// Simulate different environments
const configs = [
  {
    name: 'Development (no VITE_API_URL)',
    env: { MODE: 'development', VITE_API_URL: undefined }
  },
  {
    name: 'Development (with VITE_API_URL)',
    env: { MODE: 'development', VITE_API_URL: 'http://localhost:8080' }
  },
  {
    name: 'Production (no VITE_API_URL) - Expected for deployment',
    env: { MODE: 'production', VITE_API_URL: undefined }
  },
  {
    name: 'Production (empty VITE_API_URL) - Expected for deployment',
    env: { MODE: 'production', VITE_API_URL: '' }
  },
  {
    name: 'Production (with VITE_API_URL)',
    env: { MODE: 'production', VITE_API_URL: 'https://api.veidly.com' }
  }
];

configs.forEach(({ name, env }) => {
  // Simulate import.meta.env
  const import_meta = { env };

  // Replicate config.ts logic
  const BASE_URL = import_meta.env.VITE_API_URL || (
    import_meta.env.MODE === 'production' ? '' : 'http://localhost:8080'
  );
  const API_BASE_URL = `${BASE_URL}/api`;
  const API_BASE_URL_ROOT = BASE_URL || '';

  console.log(`📋 ${name}`);
  console.log(`   MODE: ${env.MODE}`);
  console.log(`   VITE_API_URL: ${env.VITE_API_URL === undefined ? '(undefined)' : `"${env.VITE_API_URL}"`}`);
  console.log(`   → BASE_URL: ${BASE_URL === '' ? '(empty - relative)' : BASE_URL}`);
  console.log(`   → API_BASE_URL: ${API_BASE_URL}`);
  console.log(`   → API_BASE_URL_ROOT: ${API_BASE_URL_ROOT === '' ? '(empty - relative)' : API_BASE_URL_ROOT}`);
  console.log('');
});

console.log('✅ Expected behavior for production deployment:');
console.log('   When VITE_API_URL is not set or empty:');
console.log('   - API_BASE_URL should be "/api"');
console.log('   - This makes requests to https://veidly.com/api/*');
console.log('   - Nginx proxies /api/* to backend on localhost:8080');
