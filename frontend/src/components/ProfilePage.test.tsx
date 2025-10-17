import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '../test/testUtils'
import ProfilePage from './ProfilePage'
import * as AuthContext from '../AuthContext'
import axios from 'axios'
import { mockUser, mockEvents } from '../test/mockData'

vi.mock('axios')

vi.mock('../AuthContext', () => ({
  useAuth: vi.fn(),
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({}),
  }
})

describe('ProfilePage Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockNavigate.mockClear()

    vi.mocked(AuthContext.useAuth).mockReturnValue({
      user: mockUser,
      token: 'fake-token',
      isAuthenticated: true,
      isAdmin: false,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    })
  })

  describe('Loading State', () => {
    it('should show loading state while fetching user data', () => {
      vi.mocked(axios.get).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      )

      render(<ProfilePage />)

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })
  })

  describe('Profile Display', () => {
    it('should display user profile information', async () => {
      vi.mocked(axios.get).mockResolvedValue({
        data: {
          user: mockUser,
          created_events: mockEvents,
          joined_events: [],
          past_events: []
        },
      })

      render(<ProfilePage />)

      await waitFor(() => {
        expect(screen.getByText(mockUser.name)).toBeInTheDocument()
      })
    })

    it('should display user email', async () => {
      vi.mocked(axios.get).mockResolvedValue({
        data: {
          user: mockUser,
          created_events: mockEvents,
          joined_events: [],
          past_events: []
        },
      })

      render(<ProfilePage />)

      await waitFor(() => {
        expect(screen.getByText(mockUser.email)).toBeInTheDocument()
      })
    })

    it('should display user bio when available', async () => {
      vi.mocked(axios.get).mockResolvedValue({
        data: {
          user: mockUser,
          created_events: mockEvents,
          joined_events: [],
          past_events: []
        },
      })

      render(<ProfilePage />)

      await waitFor(() => {
        if (mockUser.bio) {
          expect(screen.getByText(mockUser.bio)).toBeInTheDocument()
        }
      })
    })

    it('should display user created events', async () => {
      vi.mocked(axios.get).mockResolvedValue({
        data: {
          user: mockUser,
          created_events: mockEvents,
          joined_events: [],
          past_events: []
        },
      })

      render(<ProfilePage />)

      // Wait for accordion header to be rendered
      await waitFor(() => {
        expect(screen.getByText(/Events You're Hosting/i)).toBeInTheDocument()
      })

      // Click to expand the accordion
      const header = screen.getByText(/Events You're Hosting/i)
      header.click()

      // Now the event should be visible
      await waitFor(() => {
        expect(screen.getByText(mockEvents[0].title)).toBeInTheDocument()
      })
    })

    it('should display events in categories (created, joined, past)', async () => {
      const createdEvents = [mockEvents[0]]
      const joinedEvents = [{ ...mockEvents[1], id: 999 }]
      const pastEvents = [{ ...mockEvents[2], id: 888, is_creator: true }]

      vi.mocked(axios.get).mockResolvedValue({
        data: {
          user: mockUser,
          created_events: createdEvents,
          joined_events: joinedEvents,
          past_events: pastEvents
        },
      })

      render(<ProfilePage />)

      await waitFor(() => {
        expect(screen.getByText(/Events You're Hosting/i)).toBeInTheDocument()
        expect(screen.getByText(/Events You've Joined/i)).toBeInTheDocument()
        expect(screen.getByText(/Past Events/i)).toBeInTheDocument()
      })
    })
  })

  describe('Navigation', () => {
    it('should render back to map button', async () => {
      vi.mocked(axios.get).mockResolvedValue({
        data: {
          user: mockUser,
          created_events: [],
          joined_events: [],
          past_events: []
        },
      })

      render(<ProfilePage />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /back to map/i })).toBeInTheDocument()
      })
    })
  })

  describe('Error Handling', () => {
    it('should handle API errors gracefully', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      vi.mocked(axios.get).mockRejectedValue(new Error('API Error'))

      render(<ProfilePage />)

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalled()
      })

      consoleErrorSpy.mockRestore()
    })
  })
})
