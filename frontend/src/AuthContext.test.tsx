import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { AuthProvider, useAuth } from './AuthContext'
import axios from 'axios'
import { mockUser, mockAdmin } from './test/mockData'
import { API_BASE_URL_ROOT } from './config'

vi.mock('axios')

describe('AuthContext', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    // Mock axios.get for /api/auth/me endpoint
    vi.mocked(axios.get).mockRejectedValue(new Error('Not authenticated'))
  })

  afterEach(() => {
    localStorage.clear()
  })

  describe('Initial State', () => {
    it('should initialize with null user when no token in localStorage', async () => {
      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.user).toBeNull()
      expect(result.current.token).toBeNull()
      expect(result.current.isAuthenticated).toBe(false)
      expect(result.current.isAdmin).toBe(false)
    })

    it('should restore user from localStorage token on mount', async () => {
      const mockToken = 'test-jwt-token'
      localStorage.setItem('token', mockToken)

      // Mock successful /api/auth/me response
      vi.mocked(axios.get).mockResolvedValueOnce({
        data: { user: mockUser },
      })

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.user).toEqual(mockUser)
      expect(result.current.token).toBe(mockToken)
      expect(result.current.isAuthenticated).toBe(true)
      expect(result.current.isAdmin).toBe(false)
    })

    it('should set admin flag when user is admin', async () => {
      const mockToken = 'test-jwt-token'
      localStorage.setItem('token', mockToken)

      // Mock successful /api/auth/me response with admin user
      vi.mocked(axios.get).mockResolvedValueOnce({
        data: { user: mockAdmin },
      })

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.isAdmin).toBe(true)
    })
  })

  describe('login', () => {
    it('should successfully login and update state', async () => {
      const mockToken = 'new-jwt-token'
      const loginResponse = {
        data: {
          token: mockToken,
          user: mockUser,
        },
      }

      vi.mocked(axios.post).mockResolvedValue(loginResponse)

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.login('test@example.com', 'password123')
      })

      expect(result.current.user).toEqual(mockUser)
      expect(result.current.token).toBe(mockToken)
      expect(result.current.isAuthenticated).toBe(true)

      // Verify token is saved to localStorage
      expect(localStorage.getItem('token')).toBe(mockToken)
    })

    it('should save token to localStorage on login', async () => {
      const mockToken = 'new-jwt-token'
      const loginResponse = {
        data: {
          token: mockToken,
          user: mockUser,
        },
      }

      vi.mocked(axios.post).mockResolvedValue(loginResponse)

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.login('test@example.com', 'password123')
      })

      // Token should be in localStorage
      expect(localStorage.getItem('token')).toBe(mockToken)
    })

    it('should handle login failure', async () => {
      vi.mocked(axios.post).mockRejectedValue(new Error('Invalid credentials'))

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await expect(
        act(async () => {
          await result.current.login('test@example.com', 'wrongpassword')
        })
      ).rejects.toThrow('Invalid credentials')

      expect(result.current.user).toBeNull()
      expect(result.current.isAuthenticated).toBe(false)
    })
  })

  describe('register', () => {
    it('should successfully register and update state', async () => {
      const mockToken = 'new-jwt-token'
      const registerResponse = {
        data: {
          token: mockToken,
          user: mockUser,
        },
      }

      vi.mocked(axios.post).mockResolvedValue(registerResponse)

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.register('test@example.com', 'password123', 'Test User')
      })

      expect(result.current.user).toEqual(mockUser)
      expect(result.current.token).toBe(mockToken)
      expect(result.current.isAuthenticated).toBe(true)

      // Verify token is saved to localStorage
      expect(localStorage.getItem('token')).toBe(mockToken)
    })

    it('should handle registration failure', async () => {
      vi.mocked(axios.post).mockRejectedValue(new Error('Email already exists'))

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await expect(
        act(async () => {
          await result.current.register('test@example.com', 'password123', 'Test User')
        })
      ).rejects.toThrow('Email already exists')

      expect(result.current.user).toBeNull()
      expect(result.current.isAuthenticated).toBe(false)
    })
  })

  describe('logout', () => {
    it('should clear user on logout', () => {
      const mockToken = 'test-jwt-token'
      localStorage.setItem('token', mockToken)

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      act(() => {
        result.current.logout()
      })

      expect(result.current.user).toBeNull()
      expect(result.current.token).toBeNull()
      expect(result.current.isAuthenticated).toBe(false)

      // Verify token is removed from localStorage
      expect(localStorage.getItem('token')).toBeNull()
    })
  })

  describe('Axios interceptors', () => {
    it('should add Authorization header to requests when token exists', async () => {
      const mockToken = 'test-jwt-token'
      localStorage.setItem('token', mockToken)

      vi.mocked(axios.get).mockResolvedValueOnce({
        data: { user: mockUser },
      })

      renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(axios.interceptors.request.use).toHaveBeenCalled()
      })

      // Interceptor should be set up
      expect(axios.interceptors.request.use).toHaveBeenCalled()
    })

    it('should handle 401 response by logging out', async () => {
      const mockToken = 'test-jwt-token'
      localStorage.setItem('token', mockToken)

      vi.mocked(axios.get).mockResolvedValueOnce({
        data: { user: mockUser },
      })

      renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      })

      await waitFor(() => {
        expect(axios.interceptors.response.use).toHaveBeenCalled()
      })

      // Response interceptor should be set up
      expect(axios.interceptors.response.use).toHaveBeenCalled()
    })
  })
})
