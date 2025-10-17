import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
// Note: App already includes a Router. Don't wrap with another Router in tests.
import App from './App'
import * as AuthContext from './AuthContext'

vi.mock('./AuthContext', () => ({
  useAuth: vi.fn(),
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

vi.mock('./components/LandingPage', () => ({
  default: () => <div>Landing Page</div>,
}))

vi.mock('./components/MapView', () => ({
  default: () => <div>Map View</div>,
}))

vi.mock('./components/AdminPanel', () => ({
  default: () => <div>Admin Panel</div>,
}))

vi.mock('./components/ProfilePage', () => ({
  default: () => <div>Profile Page</div>,
}))

vi.mock('./components/PublicEventPage', () => ({
  default: () => <div>Public Event Page</div>,
}))

describe('App Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Routes - Unauthenticated User', () => {
    beforeEach(() => {
      vi.mocked(AuthContext.useAuth).mockReturnValue({
        user: null,
        token: null,
        isAuthenticated: false,
        isAdmin: false,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
      })
    })

    it('should render landing page at root path', () => {
      window.history.pushState({}, '', '/')
      render(<App />)

      expect(screen.getByText('Landing Page')).toBeInTheDocument()
    })

    it('should render map view at /map', () => {
      window.history.pushState({}, '', '/map')
      render(<App />)

      expect(screen.getByText('Map View')).toBeInTheDocument()
    })

    it('should render public event page at /event/:slug', () => {
      window.history.pushState({}, '', '/event/test-event-123')
      render(<App />)

      expect(screen.getByText('Public Event Page')).toBeInTheDocument()
    })

    it('should redirect to /map when accessing /profile without authentication', () => {
      window.history.pushState({}, '', '/profile')
      render(<App />)

      // Should redirect to map view
      expect(screen.getByText('Map View')).toBeInTheDocument()
      expect(screen.queryByText('Profile Page')).not.toBeInTheDocument()
    })

    it('should redirect to /map when accessing /admin without authentication', () => {
      window.history.pushState({}, '', '/admin')
      render(<App />)

      // Should redirect to map view
      expect(screen.getByText('Map View')).toBeInTheDocument()
      expect(screen.queryByText('Admin Panel')).not.toBeInTheDocument()
    })

    it('should render public profile page at /profile/:userId', () => {
      window.history.pushState({}, '', '/profile/123')
      render(<App />)

      expect(screen.getByText('Profile Page')).toBeInTheDocument()
    })
  })

  describe('Routes - Authenticated Regular User', () => {
    beforeEach(() => {
      vi.mocked(AuthContext.useAuth).mockReturnValue({
        user: { id: 1, email: 'user@example.com', name: 'User' } as any,
        token: 'fake-token',
        isAuthenticated: true,
        isAdmin: false,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
      })
    })

    it('should render profile page when authenticated and accessing /profile', () => {
      window.history.pushState({}, '', '/profile')
      render(<App />)

      expect(screen.getByText('Profile Page')).toBeInTheDocument()
    })

    it('should redirect to /map when non-admin tries to access /admin', () => {
      window.history.pushState({}, '', '/admin')
      render(<App />)

      // Should redirect to map view for non-admin
      expect(screen.getByText('Map View')).toBeInTheDocument()
      expect(screen.queryByText('Admin Panel')).not.toBeInTheDocument()
    })
  })

  describe('Routes - Authenticated Admin User', () => {
    beforeEach(() => {
      vi.mocked(AuthContext.useAuth).mockReturnValue({
        user: { id: 1, email: 'admin@example.com', name: 'Admin', is_admin: true } as any,
        token: 'fake-token',
        isAuthenticated: true,
        isAdmin: true,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
      })
    })

    it('should render admin panel when admin accesses /admin', () => {
      window.history.pushState({}, '', '/admin')
      render(<App />)

      expect(screen.getByText('Admin Panel')).toBeInTheDocument()
    })

    it('should render profile page when admin accesses /profile', () => {
      window.history.pushState({}, '', '/profile')
      render(<App />)

      expect(screen.getByText('Profile Page')).toBeInTheDocument()
    })
  })

  describe('ProtectedRoute Component', () => {
    it('should protect profile route from unauthenticated access', () => {
      vi.mocked(AuthContext.useAuth).mockReturnValue({
        user: null,
        token: null,
        isAuthenticated: false,
        isAdmin: false,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
      })

      window.history.pushState({}, '', '/profile')
      render(<App />)

      expect(screen.queryByText('Profile Page')).not.toBeInTheDocument()
      expect(screen.getByText('Map View')).toBeInTheDocument()
    })

    it('should protect admin route from non-admin access', () => {
      vi.mocked(AuthContext.useAuth).mockReturnValue({
        user: { id: 1, email: 'user@example.com', name: 'User' } as any,
        token: 'fake-token',
        isAuthenticated: true,
        isAdmin: false,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
      })

      window.history.pushState({}, '', '/admin')
      render(<App />)

      expect(screen.queryByText('Admin Panel')).not.toBeInTheDocument()
      expect(screen.getByText('Map View')).toBeInTheDocument()
    })
  })
})
