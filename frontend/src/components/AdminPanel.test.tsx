import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '../test/testUtils'
import AdminPanel from './AdminPanel'
import * as AuthContext from '../AuthContext'
import axios from 'axios'
import { mockUser, mockAdmin, mockEvents } from '../test/mockData'
import { API_BASE_URL } from '../config'

vi.mock('axios')

vi.mock('../AuthContext', () => ({
  useAuth: vi.fn(),
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockShowToast = vi.fn()
vi.mock('./ToastContainer', () => ({
  default: () => null,
  showToast: (message: string, type: string) => mockShowToast(message, type),
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('AdminPanel Component', () => {
  const mockLogout = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    mockNavigate.mockClear()
    mockShowToast.mockClear()

    vi.mocked(AuthContext.useAuth).mockReturnValue({
      user: mockAdmin,
      token: 'admin-token',
      isAuthenticated: true,
      isAdmin: true,
      login: vi.fn(),
      register: vi.fn(),
      logout: mockLogout,
    })
  })

  describe('Rendering', () => {
    it('should render admin panel', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByText('Admin Panel')).toBeInTheDocument()
      })
    })

    it('should render navbar with logo and navigation buttons', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByText('Veidly Admin')).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /back to map/i })).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument()
      })
    })

    it('should display admin user name', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByText(mockAdmin.name)).toBeInTheDocument()
      })
    })

    it('should render tabs for users and events management', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /users management/i })).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /events management/i })).toBeInTheDocument()
      })
    })
  })

  describe('Users Tab', () => {
    it('should display users list in users tab by default', async () => {
      const users = [mockUser, mockAdmin]
      vi.mocked(axios.get).mockResolvedValue({ data: users })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByText(`All Users (${users.length})`)).toBeInTheDocument()
        expect(screen.getByText(mockUser.name)).toBeInTheDocument()
      })
    })

    it('should display user information in table', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByText(mockUser.name)).toBeInTheDocument()
        expect(screen.getByText(mockUser.email)).toBeInTheDocument()
      })
    })

    it('should show block button for non-admin active users', async () => {
      const regularUser = { ...mockUser, is_blocked: false, is_admin: false }
      vi.mocked(axios.get).mockResolvedValue({ data: [regularUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /block/i })).toBeInTheDocument()
      })
    })

    it('should show unblock button for blocked users', async () => {
      const blockedUser = { ...mockUser, is_blocked: true, is_admin: false }
      vi.mocked(axios.get).mockResolvedValue({ data: [blockedUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /unblock/i })).toBeInTheDocument()
      })
    })

    it('should not show action buttons for admin users', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockAdmin] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.queryByRole('button', { name: /^block$/i })).not.toBeInTheDocument()
        expect(screen.queryByRole('button', { name: /^unblock$/i })).not.toBeInTheDocument()
      })
    })

    it('should call API to block user when block button is clicked', async () => {
      const regularUser = { ...mockUser, is_blocked: false, is_admin: false }
      vi.mocked(axios.get).mockResolvedValue({ data: [regularUser] })
      vi.mocked(axios.put).mockResolvedValue({})

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /block/i })).toBeInTheDocument()
      })

      const blockButton = screen.getByRole('button', { name: /block/i })
      fireEvent.click(blockButton)

      await waitFor(() => {
        expect(axios.put).toHaveBeenCalledWith(
          `${API_BASE_URL}/admin/users/${regularUser.id}/block`
        )
      })
    })

    it('should call API to unblock user when unblock button is clicked', async () => {
      const blockedUser = { ...mockUser, is_blocked: true, is_admin: false }
      vi.mocked(axios.get).mockResolvedValue({ data: [blockedUser] })
      vi.mocked(axios.put).mockResolvedValue({})

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /unblock/i })).toBeInTheDocument()
      })

      const unblockButton = screen.getByRole('button', { name: /unblock/i })
      fireEvent.click(unblockButton)

      await waitFor(() => {
        expect(axios.put).toHaveBeenCalledWith(
          `${API_BASE_URL}/admin/users/${blockedUser.id}/unblock`
        )
      })
    })

    it('should display user status badges', async () => {
      const users = [
        { ...mockUser, is_blocked: false },
        { ...mockUser, id: 2, is_blocked: true },
      ]
      vi.mocked(axios.get).mockResolvedValue({ data: users })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByText('Active')).toBeInTheDocument()
        expect(screen.getByText('Blocked')).toBeInTheDocument()
      })
    })
  })

  describe('Events Tab', () => {
    it('should switch to events tab when clicked', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /events management/i })).toBeInTheDocument()
      })

      const eventsTab = screen.getByRole('button', { name: /events management/i })

      vi.mocked(axios.get).mockResolvedValue({ data: mockEvents })
      fireEvent.click(eventsTab)

      await waitFor(() => {
        expect(screen.getByText(`All Events (${mockEvents.length})`)).toBeInTheDocument()
      })
    })

    it('should display events list in events tab', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockEvents })

      render(<AdminPanel />)

      const eventsTab = screen.getByRole('button', { name: /events management/i })
      fireEvent.click(eventsTab)

      await waitFor(() => {
        mockEvents.forEach(event => {
          expect(screen.getByText(event.title)).toBeInTheDocument()
        })
      })
    })

    it('should display event information in table', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockEvents[0]] })

      render(<AdminPanel />)

      const eventsTab = screen.getByRole('button', { name: /events management/i })
      fireEvent.click(eventsTab)

      await waitFor(() => {
        expect(screen.getByText(mockEvents[0].title)).toBeInTheDocument()
        expect(screen.getByText(mockEvents[0].creator_name)).toBeInTheDocument()
      })
    })

    it('should show delete button for each event', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockEvents[0]] })

      render(<AdminPanel />)

      const eventsTab = screen.getByRole('button', { name: /events management/i })
      fireEvent.click(eventsTab)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
      })
    })

    it('should show confirmation before deleting event', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
      vi.mocked(axios.get).mockResolvedValue({ data: [mockEvents[0]] })

      render(<AdminPanel />)

      const eventsTab = screen.getByRole('button', { name: /events management/i })
      fireEvent.click(eventsTab)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
      })

      const deleteButton = screen.getByRole('button', { name: /delete/i })
      fireEvent.click(deleteButton)

      expect(confirmSpy).toHaveBeenCalled()
      confirmSpy.mockRestore()
    })

    it('should call API to delete event when confirmed', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
      vi.mocked(axios.get).mockResolvedValue({ data: [mockEvents[0]] })
      vi.mocked(axios.delete).mockResolvedValue({})

      render(<AdminPanel />)

      const eventsTab = screen.getByRole('button', { name: /events management/i })
      fireEvent.click(eventsTab)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
      })

      const deleteButton = screen.getByRole('button', { name: /delete/i })
      fireEvent.click(deleteButton)

      await waitFor(() => {
        expect(axios.delete).toHaveBeenCalledWith(
          `${API_BASE_URL}/admin/events/${mockEvents[0].id}`
        )
      })

      confirmSpy.mockRestore()
    })

    it('should not call API to delete event when cancelled', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
      vi.mocked(axios.get).mockResolvedValue({ data: [mockEvents[0]] })

      render(<AdminPanel />)

      const eventsTab = screen.getByRole('button', { name: /events management/i })
      fireEvent.click(eventsTab)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
      })

      const deleteButton = screen.getByRole('button', { name: /delete/i })
      fireEvent.click(deleteButton)

      expect(axios.delete).not.toHaveBeenCalled()
      confirmSpy.mockRestore()
    })

    it('should display empty state when no events', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [] })

      render(<AdminPanel />)

      const eventsTab = screen.getByRole('button', { name: /events management/i })
      fireEvent.click(eventsTab)

      await waitFor(() => {
        expect(screen.getByText(/no events found/i)).toBeInTheDocument()
      })
    })
  })

  describe('Navigation', () => {
    it('should navigate to map when "Back to Map" is clicked', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /back to map/i })).toBeInTheDocument()
      })

      const backButton = screen.getByRole('button', { name: /back to map/i })
      fireEvent.click(backButton)

      expect(mockNavigate).toHaveBeenCalledWith('/map')
    })

    it('should navigate to home when logo is clicked', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByText('Veidly Admin')).toBeInTheDocument()
      })

      const logo = screen.getByText('Veidly Admin')
      fireEvent.click(logo)

      expect(mockNavigate).toHaveBeenCalledWith('/')
    })

    it('should call logout when logout button is clicked', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [mockUser] })

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument()
      })

      const logoutButton = screen.getByRole('button', { name: /logout/i })
      fireEvent.click(logoutButton)

      expect(mockLogout).toHaveBeenCalled()
    })
  })

  describe('Loading State', () => {
    it('should show loading state while fetching data', async () => {
      vi.mocked(axios.get).mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve({ data: [mockUser] }), 100))
      )

      render(<AdminPanel />)

      expect(screen.getByText(/loading/i)).toBeInTheDocument()

      await waitFor(() => {
        expect(screen.queryByText(/loading/i)).not.toBeInTheDocument()
      })
    })
  })

  describe('Error Handling', () => {
    it('should handle API errors when fetching users', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      vi.mocked(axios.get).mockRejectedValue(new Error('API Error'))

      render(<AdminPanel />)

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalled()
      })

      consoleErrorSpy.mockRestore()
    })

    it('should show toast on block user failure', async () => {
      const regularUser = { ...mockUser, is_blocked: false, is_admin: false }
      vi.mocked(axios.get).mockResolvedValue({ data: [regularUser] })
      vi.mocked(axios.put).mockRejectedValue(new Error('API Error'))

      render(<AdminPanel />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /block/i })).toBeInTheDocument()
      })

      const blockButton = screen.getByRole('button', { name: /block/i })
      fireEvent.click(blockButton)

      await waitFor(() => {
        expect(mockShowToast).toHaveBeenCalledWith('Failed to block user', 'error')
      })
    })

    it('should show toast on delete event failure', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
      vi.mocked(axios.get).mockResolvedValue({ data: [mockEvents[0]] })
      vi.mocked(axios.delete).mockRejectedValue(new Error('API Error'))

      render(<AdminPanel />)

      const eventsTab = screen.getByRole('button', { name: /events management/i })
      fireEvent.click(eventsTab)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
      })

      const deleteButton = screen.getByRole('button', { name: /delete/i })
      fireEvent.click(deleteButton)

      await waitFor(() => {
        expect(mockShowToast).toHaveBeenCalledWith('Failed to delete event', 'error')
      })

      confirmSpy.mockRestore()
    })
  })
})
