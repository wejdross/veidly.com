import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '../test/testUtils'
import ParticipantModal from './ParticipantModal'
import { mockUser } from '../test/mockData'
import * as api from '../api'

vi.mock('../api', () => ({
  api: {
    getEventParticipants: vi.fn(),
  },
}))

describe('ParticipantModal Component', () => {
  const mockOnClose = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Rendering', () => {
    it('should render modal', () => {
      vi.mocked(api.api.getEventParticipants).mockResolvedValue([])

      render(
        <ParticipantModal
          eventId={1}
          eventTitle="Coffee & Chat"
          onClose={mockOnClose}
        />
      )

      // Modal renders - check for title
      expect(screen.getByText(/coffee & chat/i)).toBeInTheDocument()
    })

    it('should show loading state initially', () => {
      vi.mocked(api.api.getEventParticipants).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      )

      render(
        <ParticipantModal
          eventId={1}
          eventTitle="Coffee & Chat"
          onClose={mockOnClose}
        />
      )

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('should display participants after loading', async () => {
      const participants = [mockUser]
      vi.mocked(api.api.getEventParticipants).mockResolvedValue(participants)

      render(
        <ParticipantModal
          eventId={1}
          eventTitle="Coffee & Chat"
          onClose={mockOnClose}
        />
      )

      await waitFor(() => {
        expect(screen.getByText(mockUser.name)).toBeInTheDocument()
      })
    })

    it('should show message when no participants', async () => {
      vi.mocked(api.api.getEventParticipants).mockResolvedValue([])

      render(
        <ParticipantModal
          eventId={1}
          eventTitle="Coffee & Chat"
          onClose={mockOnClose}
        />
      )

      await waitFor(() => {
        expect(screen.getByText(/no participants yet/i)).toBeInTheDocument()
      })
    })

    it('should show error message on API failure', async () => {
      vi.mocked(api.api.getEventParticipants).mockRejectedValue(new Error('API Error'))

      render(
        <ParticipantModal
          eventId={1}
          eventTitle="Coffee & Chat"
          onClose={mockOnClose}
        />
      )

      await waitFor(() => {
        expect(screen.getByText(/failed to load participants/i)).toBeInTheDocument()
      })
    })
  })

  describe('Participant Display', () => {
    it('should display participant names', async () => {
      const participants = [mockUser, { ...mockUser, id: 2, name: 'Jane Doe' }]
      vi.mocked(api.api.getEventParticipants).mockResolvedValue(participants)

      render(
        <ParticipantModal
          eventId={1}
          eventTitle="Coffee & Chat"
          onClose={mockOnClose}
        />
      )

      await waitFor(() => {
        expect(screen.getByText('Test User')).toBeInTheDocument()
        expect(screen.getByText('Jane Doe')).toBeInTheDocument()
      })
    })

    it('should display multiple participant names', async () => {
      const participants = [mockUser, { ...mockUser, id: 2, name: 'Jane Doe' }]
      vi.mocked(api.api.getEventParticipants).mockResolvedValue(participants)

      render(
        <ParticipantModal
          eventId={1}
          eventTitle="Coffee & Chat"
          onClose={mockOnClose}
        />
      )

      await waitFor(() => {
        expect(screen.getByText('Jane Doe')).toBeInTheDocument()
      })
    })
  })

  describe('Modal Interactions', () => {
    it('should call onClose when close button is clicked', async () => {
      vi.mocked(api.api.getEventParticipants).mockResolvedValue([])

      render(
        <ParticipantModal
          eventId={1}
          eventTitle="Coffee & Chat"
          onClose={mockOnClose}
        />
      )

      await waitFor(() => {
        expect(screen.getByText(/no participants yet/i)).toBeInTheDocument()
      })

      const closeButtons = screen.getAllByRole('button')
      const closeButton = closeButtons[0]
      fireEvent.click(closeButton)

      expect(mockOnClose).toHaveBeenCalled()
    })

    it('should call onClose when overlay is clicked', async () => {
      vi.mocked(api.api.getEventParticipants).mockResolvedValue([])

      render(
        <ParticipantModal
          eventId={1}
          eventTitle="Coffee & Chat"
          onClose={mockOnClose}
        />
      )

      await waitFor(() => {
        expect(screen.getByText(/no participants yet/i)).toBeInTheDocument()
      })

      const overlay = document.querySelector('.modal-overlay')
      if (overlay) {
        fireEvent.click(overlay)
      }

      expect(mockOnClose).toHaveBeenCalled()
    })
  })
})
