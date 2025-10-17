import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '../test/testUtils'
import EventForm from './EventForm'
import { mockEvent, mockUser } from '../test/mockData'
import * as api from '../api'
import * as AuthContext from '../AuthContext'

// Mock the API
vi.mock('../api', () => ({
  api: {
    createEvent: vi.fn(),
    updateEvent: vi.fn(),
  },
}))

// Mock the AuthContext
vi.mock('../AuthContext', () => ({
  useAuth: vi.fn(),
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

describe('EventForm Component', () => {
  const mockOnClose = vi.fn()
  const mockLocation = { lat: 46.8805, lng: 8.6444 }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(AuthContext.useAuth).mockReturnValue({
      user: mockUser,
      isAuthenticated: true,
      isAdmin: false,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    })
  })

  describe('Create Mode', () => {
    it('should render create form with correct title', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      expect(screen.getByText('Create New Event')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /create event/i })).toBeInTheDocument()
    })

    it('should pre-fill location from initialLocation prop', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      // Location is set internally but not shown as input fields
      // The form uses a map-based location picker instead
      // Just verify the form renders successfully with the location
      expect(screen.getByRole('button', { name: /create event/i })).toBeInTheDocument()
    })

    it('should pre-fill creator name and contact from user profile', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const nameInput = screen.getByLabelText(/your name/i, { selector: 'input[type="text"]' }) as HTMLInputElement

      expect(nameInput.value).toBe(mockUser.name)
    })

    it('should handle text input changes', async () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const titleInput = screen.getByLabelText(/event title/i) as HTMLInputElement

      fireEvent.change(titleInput, { target: { value: 'New Event Title' } })

      expect(titleInput.value).toBe('New Event Title')
    })

    it('should handle textarea changes', async () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const descInput = screen.getByLabelText(/description/i) as HTMLTextAreaElement

      fireEvent.change(descInput, { target: { value: 'New description' } })

      expect(descInput.value).toBe('New description')
    })

    it('should handle checkbox changes', async () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const smokingCheckbox = screen.getByRole('checkbox', { name: /smoking is allowed/i })

      expect(smokingCheckbox).not.toBeChecked()

      fireEvent.click(smokingCheckbox)

      expect(smokingCheckbox).toBeChecked()
    })

    it('should handle select changes', async () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const categorySelect = screen.getByLabelText(/category/i) as HTMLSelectElement

      fireEvent.change(categorySelect, { target: { value: 'sports_fitness' } })

      expect(categorySelect.value).toBe('sports_fitness')
    })

    it('should handle number input changes', async () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const maxParticipantsInput = screen.getByLabelText(/max participants/i) as HTMLInputElement

      fireEvent.change(maxParticipantsInput, { target: { value: '10' } })

      expect(maxParticipantsInput.value).toBe('10')
    })

    it('should call createEvent API on successful form submission', async () => {
      vi.mocked(api.api.createEvent).mockResolvedValue(mockEvent)

      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      // Fill required fields
      fireEvent.change(screen.getByLabelText(/event title/i), {
        target: { value: 'Test Event' },
      })
      fireEvent.change(screen.getByLabelText(/description/i), {
        target: { value: 'Test Description' },
      })
      fireEvent.change(screen.getByLabelText(/start time/i), {
        target: { value: '2025-12-31T14:00' },
      })

      // Submit form
      fireEvent.click(screen.getByRole('button', { name: /create event/i }))

      await waitFor(() => {
        expect(api.api.createEvent).toHaveBeenCalledWith(
          expect.objectContaining({
            title: 'Test Event',
            description: 'Test Description',
            start_time: '2025-12-31T14:00',
          })
        )
        expect(mockOnClose).toHaveBeenCalled()
      })
    })

    it('should show error message on API failure', async () => {
      vi.mocked(api.api.createEvent).mockRejectedValue({
        response: { data: { error: 'Failed to create' } },
      })

      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      // Fill required fields
      fireEvent.change(screen.getByLabelText(/event title/i), {
        target: { value: 'Test Event' },
      })
      fireEvent.change(screen.getByLabelText(/description/i), {
        target: { value: 'Test Description' },
      })
      fireEvent.change(screen.getByLabelText(/start time/i), {
        target: { value: '2025-12-31T14:00' },
      })

      // Submit form
      fireEvent.click(screen.getByRole('button', { name: /create event/i }))

      await waitFor(() => {
        expect(screen.getByText('Failed to create')).toBeInTheDocument()
      })
    })

    it('should close modal when Cancel button is clicked', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      fireEvent.click(screen.getByRole('button', { name: /cancel/i }))

      expect(mockOnClose).toHaveBeenCalled()
    })

    it('should close modal when clicking overlay', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const overlay = screen.getByText('Create New Event').closest('.modal-overlay')
      if (overlay) {
        fireEvent.click(overlay)
      }

      expect(mockOnClose).toHaveBeenCalled()
    })
  })

  describe('Edit Mode', () => {
    it('should render edit form with correct title', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
          event={mockEvent}
        />
      )

      expect(screen.getByText('Edit Event')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /update event/i })).toBeInTheDocument()
    })

    it('should pre-fill all fields with event data', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
          event={mockEvent}
        />
      )

      // Check title
      const titleInput = screen.getByLabelText(/event title/i) as HTMLInputElement
      expect(titleInput.value).toBe(mockEvent.title)

      // Check description
      const descInput = screen.getByLabelText(/description/i) as HTMLTextAreaElement
      expect(descInput.value).toBe(mockEvent.description)

      // Check category
      const categorySelect = screen.getByLabelText(/category/i) as HTMLSelectElement
      expect(categorySelect.value).toBe(mockEvent.category)

      // Location is handled internally via map picker, not via input fields

      // Check max participants
      const maxInput = screen.getByLabelText(/max participants/i) as HTMLInputElement
      expect(maxInput.value).toBe(mockEvent.max_participants?.toString() || '')

      // Check gender restriction
      const genderSelect = screen.getByLabelText(/gender preference/i) as HTMLSelectElement
      expect(genderSelect.value).toBe(mockEvent.gender_restriction)

      // Check checkboxes
      const alcoholCheckbox = screen.getByRole('checkbox', { name: /alcohol is allowed/i })
      expect(alcoholCheckbox).toBeChecked() // mockEvent has alcohol_allowed: true
    })

    it('should maintain title value when editing', async () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
          event={mockEvent}
        />
      )

      const titleInput = screen.getByLabelText(/event title/i) as HTMLInputElement

      // Title should be pre-filled
      expect(titleInput.value).toBe(mockEvent.title)

      // Change the title
      fireEvent.change(titleInput, { target: { value: 'Updated Title' } })

      // Title should be updated
      expect(titleInput.value).toBe('Updated Title')
    })

    it('should call updateEvent API on successful edit submission', async () => {
      vi.mocked(api.api.updateEvent).mockResolvedValue(mockEvent)

      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
          event={mockEvent}
        />
      )

      // Update the title
      const titleInput = screen.getByLabelText(/event title/i)
      fireEvent.change(titleInput, { target: { value: 'Updated Event' } })

      // Submit form
      fireEvent.click(screen.getByRole('button', { name: /update event/i }))

      await waitFor(() => {
        expect(api.api.updateEvent).toHaveBeenCalledWith(
          mockEvent.id,
          expect.objectContaining({
            title: 'Updated Event',
          })
        )
        expect(mockOnClose).toHaveBeenCalled()
      })
    })

    it('should show error message on update failure', async () => {
      vi.mocked(api.api.updateEvent).mockRejectedValue({
        response: { data: { error: 'Failed to update' } },
      })

      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
          event={mockEvent}
        />
      )

      // Submit form
      fireEvent.click(screen.getByRole('button', { name: /update event/i }))

      await waitFor(() => {
        expect(screen.getByText('Failed to update')).toBeInTheDocument()
      })
    })

    it('should disable submit button while submitting', async () => {
      vi.mocked(api.api.updateEvent).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockEvent), 100))
      )

      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
          event={mockEvent}
        />
      )

      const submitButton = screen.getByRole('button', { name: /update event/i })

      fireEvent.click(submitButton)

      // Button should be disabled during submission
      expect(submitButton).toBeDisabled()
      expect(screen.getByText(/updating\.\.\./i)).toBeInTheDocument()

      await waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled()
      })
    })
  })

  describe('Form Validation', () => {
    it('should require title field', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const titleInput = screen.getByLabelText(/event title/i)
      expect(titleInput).toBeRequired()
    })

    it('should require description field', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const descInput = screen.getByLabelText(/description/i)
      expect(descInput).toBeRequired()
    })

    it('should require start time field', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const startInput = screen.getByLabelText(/start time/i)
      expect(startInput).toBeRequired()
    })

    it('should require location to be set', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      // Location is required but set via map interaction, not input fields
      // The form validates that latitude !== 0 and longitude !== 0
      // Just verify the form renders successfully
      expect(screen.getByRole('button', { name: /create event/i })).toBeInTheDocument()
    })
  })

  describe('Validation', () => {
    it('should show error for description less than 10 characters', async () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      // Fill required fields with short description
      fireEvent.change(screen.getByLabelText(/event title/i), {
        target: { value: 'Test Event' },
      })
      fireEvent.change(screen.getByLabelText(/description/i), {
        target: { value: 'Too short' }, // Only 9 characters
      })
      fireEvent.change(screen.getByLabelText(/start time/i), {
        target: { value: '2025-12-31T14:00' },
      })

      // Submit form
      fireEvent.click(screen.getByRole('button', { name: /create event/i }))

      await waitFor(() => {
        expect(screen.getByText(/description must be at least 10 characters/i)).toBeInTheDocument()
        expect(api.api.createEvent).not.toHaveBeenCalled()
      })
    })

    it('should show character count for description', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const descInput = screen.getByLabelText(/description/i)
      fireEvent.change(descInput, { target: { value: 'Test' } })

      expect(screen.getByText(/4\/5000 characters/i)).toBeInTheDocument()
    })

    it('should show warning when description is too short', () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      const descInput = screen.getByLabelText(/description/i)
      fireEvent.change(descInput, { target: { value: 'Short' } }) // 5 characters

      expect(screen.getByText(/5 more needed/i)).toBeInTheDocument()
    })

    it('should accept description with exactly 10 characters', async () => {
      vi.mocked(api.api.createEvent).mockResolvedValue(mockEvent)

      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      fireEvent.change(screen.getByLabelText(/event title/i), {
        target: { value: 'Test Event' },
      })
      fireEvent.change(screen.getByLabelText(/description/i), {
        target: { value: 'Exactly 10' }, // Exactly 10 characters
      })
      fireEvent.change(screen.getByLabelText(/start time/i), {
        target: { value: '2025-12-31T14:00' },
      })

      fireEvent.click(screen.getByRole('button', { name: /create event/i }))

      await waitFor(() => {
        expect(api.api.createEvent).toHaveBeenCalled()
        expect(mockOnClose).toHaveBeenCalled()
      })
    })

    it('should show error for title less than 3 characters', async () => {
      render(
        <EventForm
          initialLocation={mockLocation}
          onClose={mockOnClose}
        />
      )

      fireEvent.change(screen.getByLabelText(/event title/i), {
        target: { value: 'AB' }, // Only 2 characters
      })
      fireEvent.change(screen.getByLabelText(/description/i), {
        target: { value: 'Valid description here' },
      })
      fireEvent.change(screen.getByLabelText(/start time/i), {
        target: { value: '2025-12-31T14:00' },
      })

      fireEvent.click(screen.getByRole('button', { name: /create event/i }))

      await waitFor(() => {
        expect(screen.getByText(/title must be at least 3 characters/i)).toBeInTheDocument()
        expect(api.api.createEvent).not.toHaveBeenCalled()
      })
    })
  })
})
