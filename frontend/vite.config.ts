import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { writeFileSync } from 'fs'
import { resolve } from 'path'

// Plugin to generate version file during build
const versionPlugin = () => ({
  name: 'version-plugin',
  buildStart() {
    const version = {
      buildTime: new Date().toISOString(),
      version: process.env.npm_package_version || '0.1.0'
    }
    // Write version.json to public folder so it's copied to dist
    writeFileSync(
      resolve(__dirname, 'public', 'version.json'),
      JSON.stringify(version, null, 2)
    )
  }
})

export default defineConfig({
  plugins: [react(), versionPlugin()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'src/test/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/mockData.ts',
        'dist/',
      ],
    },
  },
})
