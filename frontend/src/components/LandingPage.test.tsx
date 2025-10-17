import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '../test/testUtils'
import LandingPage from './LandingPage'

// Mock useNavigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('LandingPage Component', () => {
  beforeEach(() => {
    mockNavigate.mockClear()
  })

  describe('Rendering', () => {
    it('should render the landing page', () => {
      render(<LandingPage />)

      expect(screen.getByText('Veidly')).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: 'Real People, Real Connections' })).toBeInTheDocument()
    })

    it('should display hero section with main heading', () => {
      render(<LandingPage />)

      expect(screen.getByRole('heading', { name: 'Real People, Real Connections' })).toBeInTheDocument()
      expect(screen.getByText(/A safe, welcoming space/i)).toBeInTheDocument()
    })

    it('should display hero points', () => {
      render(<LandingPage />)

      expect(screen.getByText('Authentic communities')).toBeInTheDocument()
      expect(screen.getByText('Privacy-focused & safe')).toBeInTheDocument()
      expect(screen.getByText('Respectful environment')).toBeInTheDocument()
    })

    it('should display features section', () => {
      render(<LandingPage />)

      expect(screen.getByText('Why People Love Veidly')).toBeInTheDocument()
      expect(screen.getByText('Safe Event Discovery')).toBeInTheDocument()
    })

    it('should render navigation bar with logo', () => {
      render(<LandingPage />)

      const logo = screen.getAllByText('Veidly')[0]
      expect(logo).toBeInTheDocument()
    })

    it('should display "Find Events" button in navbar', () => {
      render(<LandingPage />)

      const buttons = screen.getAllByRole('button', { name: /find events/i })
      expect(buttons.length).toBeGreaterThan(0)
    })

    it('should display "Discover Events Near You" CTA button', () => {
      render(<LandingPage />)

      expect(screen.getByRole('button', { name: /discover events near you/i })).toBeInTheDocument()
    })
  })

  describe('Navigation', () => {
    it('should navigate to /map when navbar "Find Events" button is clicked', () => {
      render(<LandingPage />)

      const findEventsButtons = screen.getAllByRole('button', { name: /find events/i })
      fireEvent.click(findEventsButtons[0])

      expect(mockNavigate).toHaveBeenCalledWith('/map')
    })

    it('should navigate to /map when hero CTA button is clicked', () => {
      render(<LandingPage />)

      const exploreCTA = screen.getByRole('button', { name: /discover events near you/i })
      fireEvent.click(exploreCTA)

      expect(mockNavigate).toHaveBeenCalledWith('/map')
    })
  })

  describe('Content', () => {
    it('should display correct hero subtitle', () => {
      render(<LandingPage />)

      expect(screen.getByText(/A safe, welcoming space to discover authentic friendships/i)).toBeInTheDocument()
      expect(screen.getByText(/No fake profiles. No algorithms/i)).toBeInTheDocument()
    })

    it('should have all three hero points with icons', () => {
      render(<LandingPage />)

      // Check for the three hero points
      expect(screen.getByText('Authentic communities')).toBeInTheDocument()
      expect(screen.getByText('Privacy-focused & safe')).toBeInTheDocument()
      expect(screen.getByText('Respectful environment')).toBeInTheDocument()
    })

    it('should render feature cards', () => {
      render(<LandingPage />)

      // At least one feature card should be rendered
      expect(screen.getByText('Safe Event Discovery')).toBeInTheDocument()
      expect(screen.getByText('Privacy Controls')).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('should have accessible buttons', () => {
      render(<LandingPage />)

      const buttons = screen.getAllByRole('button')
      expect(buttons.length).toBeGreaterThan(0)

      buttons.forEach(button => {
        expect(button).toBeEnabled()
      })
    })

    it('should have heading hierarchy', () => {
      render(<LandingPage />)

      const mainHeading = screen.getByRole('heading', { level: 1 })
      expect(mainHeading).toHaveTextContent('Real People, Real Connections')

      const subheadings = screen.getAllByRole('heading', { level: 2 })
      expect(subheadings.length).toBeGreaterThan(0)
    })
  })
})
