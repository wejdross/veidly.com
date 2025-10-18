import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '../test/testUtils'
import PublicEventPage from './PublicEventPage'
import * as AuthContext from '../AuthContext'
import * as api from '../api'
import { mockEvent } from '../test/mockData'

vi.mock('../AuthContext', () => ({
  useAuth: vi.fn(),
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

vi.mock('../api', () => ({
  api: {
    getPublicEvent: vi.fn(),
    joinEvent: vi.fn(),
    leaveEvent: vi.fn(),
  },
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ slug: 'test-event-abc123' }),
  }
})

// Mock Leaflet components
vi.mock('react-leaflet', () => ({
  MapContainer: ({ children }: any) => <div data-testid="map-container">{children}</div>,
  TileLayer: () => <div data-testid="tile-layer" />,
  Marker: ({ children }: any) => <div data-testid="marker">{children}</div>,
  Popup: ({ children }: any) => <div>{children}</div>,
}))

describe('PublicEventPage Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockNavigate.mockClear()

    vi.mocked(AuthContext.useAuth).mockReturnValue({
      user: null,
      token: null,
      isAuthenticated: false,
      isAdmin: false,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    })
  })

  describe('Loading State', () => {
    it('should show loading state while fetching event', () => {
      vi.mocked(api.api.getPublicEvent).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      )

      render(<PublicEventPage />)

      expect(screen.getByText(/loading event/i)).toBeInTheDocument()
    })
  })

  describe('Event Display', () => {
    it('should display event details when loaded', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: mockEvent.title })).toBeInTheDocument()
        expect(screen.getByText(mockEvent.description)).toBeInTheDocument()
      })
    })

    it('should display event creator information', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(mockEvent.creator_name)).toBeInTheDocument()
      })
    })

    it('should display event category', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/social & drinks/i)).toBeInTheDocument()
      })
    })

    it('should display event start time', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/start:/i)).toBeInTheDocument()
      })
    })

    it('should display event end time if available', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/end:/i)).toBeInTheDocument()
      })
    })

    it('should display participant count when max_participants is set', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/participants/i)).toBeInTheDocument()
        expect(screen.getByText(new RegExp(`${mockEvent.participant_count}\\s*/\\s*${mockEvent.max_participants}\\s*joined`, 'i'))).toBeInTheDocument()
      })
    })

    it('should display event restrictions', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/smoking:/i)).toBeInTheDocument()
        expect(screen.getByText(/alcohol:/i)).toBeInTheDocument()
      })
    })

    it('should render map with event location', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByTestId('map-container')).toBeInTheDocument()
        expect(screen.getByTestId('marker')).toBeInTheDocument()
      })
    })
  })

  describe('Error Handling', () => {
    it('should display error message when event is not found', async () => {
      vi.mocked(api.api.getPublicEvent).mockRejectedValue(new Error('Not found'))

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/event not found/i)).toBeInTheDocument()
      })
    })

    it('should display error message on fetch failure', async () => {
      vi.mocked(api.api.getPublicEvent).mockRejectedValue(new Error('Network error'))

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/failed to load event/i)).toBeInTheDocument()
      })
    })

    it('should show "Browse Events" button when event not found', async () => {
      vi.mocked(api.api.getPublicEvent).mockRejectedValue(new Error('Not found'))

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /browse events/i })).toBeInTheDocument()
      })
    })
  })

  describe('Navigation - Unauthenticated', () => {
    it('should show "Browse Events" button when not authenticated', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /browse events/i })).toBeInTheDocument()
      })
    })

    it('should show "Sign In" button when not authenticated', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /^sign in$/i })).toBeInTheDocument()
      })
    })

    it('should navigate to / when "Sign In" is clicked', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /^sign in$/i })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: /^sign in$/i }))
      expect(mockNavigate).toHaveBeenCalledWith('/?returnTo=/event/test-event-abc123')
    })

    it('should show CTA to sign in when not authenticated', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/want to join this event/i)).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /sign in to join/i })).toBeInTheDocument()
      })
    })

    it('should navigate with returnTo parameter when "Sign In to Join" is clicked', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /sign in to join/i })).toBeInTheDocument()
      })

      const signInButton = screen.getByRole('button', { name: /sign in to join/i })
      fireEvent.click(signInButton)

      expect(mockNavigate).toHaveBeenCalledWith('/?returnTo=/event/test-event-abc123')
    })
  })

  describe('Navigation - Authenticated', () => {
    beforeEach(() => {
      vi.mocked(AuthContext.useAuth).mockReturnValue({
        user: { id: 1, email: 'user@example.com', name: 'User' } as any,
        token: 'fake-token',
        isAuthenticated: true,
        isAdmin: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
      })
    })

    it('should show "Go to Map" button when authenticated', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /go to map/i })).toBeInTheDocument()
      })
    })

    it('should not show sign in CTA when authenticated', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.queryByText(/want to join this event/i)).not.toBeInTheDocument()
      })
    })

    it('should navigate to /map when "Go to Map" is clicked', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /go to map/i })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: /go to map/i }))
      expect(mockNavigate).toHaveBeenCalledWith('/map')
    })
  })

  describe('Header', () => {
    it('should display Veidly logo', async () => {
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(mockEvent)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText('Veidly')).toBeInTheDocument()
      })
    })

    it('should navigate to /map when error state "Browse Events" is clicked', async () => {
      vi.mocked(api.api.getPublicEvent).mockRejectedValue(new Error('Not found'))

      render(<PublicEventPage />)

      await waitFor(() => {
        const browseButtons = screen.getAllByRole('button', { name: /browse events/i })
        fireEvent.click(browseButtons[0])
      })

      expect(mockNavigate).toHaveBeenCalledWith('/map')
    })
  })

  describe('Age Restrictions', () => {
    it('should not display age restriction when age range is default', async () => {
      const eventWithDefaultAge = { ...mockEvent, age_min: 0, age_max: 99 }
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(eventWithDefaultAge)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.queryByText(/age:/i)).not.toBeInTheDocument()
      })
    })

    it('should display age restriction when set', async () => {
      const eventWithAgeLimit = { ...mockEvent, age_min: 21, age_max: 35 }
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(eventWithAgeLimit)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/age:/i)).toBeInTheDocument()
        expect(screen.getByText(/21/)).toBeInTheDocument()
        expect(screen.getByText(/35/)).toBeInTheDocument()
      })
    })
  })

  describe('Gender Restrictions', () => {
    it('should not display gender restriction when set to "any"', async () => {
      const eventWithNoGenderRestriction = { ...mockEvent, gender_restriction: 'any' }
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(eventWithNoGenderRestriction)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.queryByText(/gender:/i)).not.toBeInTheDocument()
      })
    })

    it('should display gender restriction when set', async () => {
      const eventWithGenderRestriction = { ...mockEvent, gender_restriction: 'female' }
      vi.mocked(api.api.getPublicEvent).mockResolvedValue(eventWithGenderRestriction)

      render(<PublicEventPage />)

      await waitFor(() => {
        expect(screen.getByText(/gender:/i)).toBeInTheDocument()
        expect(screen.getByText(/female/i)).toBeInTheDocument()
      })
    })
  })
})
