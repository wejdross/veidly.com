import { expect, afterEach, beforeEach, vi } from 'vitest'
import { cleanup } from '@testing-library/react'
import * as matchers from '@testing-library/jest-dom/matchers'
import axios from 'axios'

expect.extend(matchers)

beforeEach(() => {
  vi.clearAllMocks()
})

afterEach(() => {
  cleanup()
  // Clear axios interceptors after each test
  axios.interceptors.request.handlers = []
  axios.interceptors.response.handlers = []
  // Clear localStorage
  localStorage.clear()
})

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})
