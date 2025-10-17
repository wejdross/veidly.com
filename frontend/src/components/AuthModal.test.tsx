import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '../test/testUtils'
import AuthModal from './AuthModal'
import * as AuthContext from '../AuthContext'

vi.mock('../AuthContext', () => ({
  useAuth: vi.fn(),
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

describe('AuthModal Component', () => {
  const mockOnClose = vi.fn()
  const mockLogin = vi.fn()
  const mockRegister = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(AuthContext.useAuth).mockReturnValue({
      user: null,
      isAuthenticated: false,
      isAdmin: false,
      login: mockLogin,
      register: mockRegister,
      logout: vi.fn(),
    })
  })

  describe('Login Mode', () => {
    it('should render login form by default', () => {
      render(<AuthModal onClose={mockOnClose} />)

      expect(screen.getByText('Welcome Back!')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /login/i })).toBeInTheDocument()
      expect(screen.queryByLabelText(/name/i)).not.toBeInTheDocument()
    })

    it('should handle email input', () => {
      render(<AuthModal onClose={mockOnClose} />)

      const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } })

      expect(emailInput.value).toBe('test@example.com')
    })

    it('should handle password input', () => {
      render(<AuthModal onClose={mockOnClose} />)

      const passwordInput = screen.getByLabelText(/password/i) as HTMLInputElement
      fireEvent.change(passwordInput, { target: { value: 'password123' } })

      expect(passwordInput.value).toBe('password123')
    })

    it('should call login on form submission', async () => {
      mockLogin.mockResolvedValue(undefined)

      render(<AuthModal onClose={mockOnClose} />)

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'test@example.com' },
      })
      fireEvent.change(screen.getByLabelText(/password/i), {
        target: { value: 'password123' },
      })

      fireEvent.click(screen.getByRole('button', { name: /login/i }))

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('test@example.com', 'password123')
        expect(mockOnClose).toHaveBeenCalled()
      })
    })

    it('should show error message on login failure', async () => {
      mockLogin.mockRejectedValue({
        response: { data: { error: 'Invalid credentials' } },
      })

      render(<AuthModal onClose={mockOnClose} />)

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'test@example.com' },
      })
      fireEvent.change(screen.getByLabelText(/password/i), {
        target: { value: 'wrongpassword' },
      })

      fireEvent.click(screen.getByRole('button', { name: /login/i }))

      await waitFor(() => {
        expect(screen.getByText('Invalid credentials')).toBeInTheDocument()
      })
    })

    it('should switch to register mode when clicking sign up link', () => {
      render(<AuthModal onClose={mockOnClose} />)

      const signUpLink = screen.getByRole('button', { name: /sign up/i })
      fireEvent.click(signUpLink)

      expect(screen.getByText('Join Veidly')).toBeInTheDocument()
      expect(screen.getByLabelText(/name/i)).toBeInTheDocument()
    })
  })

  describe('Register Mode', () => {
    it('should render register form when initialMode is register', () => {
      render(<AuthModal onClose={mockOnClose} initialMode="register" />)

      expect(screen.getByText('Join Veidly')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /create account/i })).toBeInTheDocument()
      expect(screen.getByLabelText(/name/i)).toBeInTheDocument()
    })

    it('should handle name input', () => {
      render(<AuthModal onClose={mockOnClose} initialMode="register" />)

      const nameInput = screen.getByLabelText(/name/i) as HTMLInputElement
      fireEvent.change(nameInput, { target: { value: 'John Doe' } })

      expect(nameInput.value).toBe('John Doe')
    })

    it('should call register on form submission', async () => {
      mockRegister.mockResolvedValue(undefined)

      render(<AuthModal onClose={mockOnClose} initialMode="register" />)

      fireEvent.change(screen.getByLabelText(/name/i), {
        target: { value: 'John Doe' },
      })
      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'john@example.com' },
      })
      fireEvent.change(screen.getByLabelText(/password/i), {
        target: { value: 'password123' },
      })

      fireEvent.click(screen.getByRole('button', { name: /create account/i }))

      await waitFor(() => {
        expect(mockRegister).toHaveBeenCalledWith('john@example.com', 'password123', 'John Doe')
        expect(mockOnClose).toHaveBeenCalled()
      })
    })

    it('should switch to login mode when switching modes', () => {
      render(<AuthModal onClose={mockOnClose} initialMode="register" />)

      // Find and click the button to switch modes
      const switchButtons = screen.getAllByRole('button')
      const loginButton = switchButtons.find(btn => btn.textContent?.includes('Log in'))

      if (loginButton) {
        fireEvent.click(loginButton)
        expect(screen.getByText('Welcome Back!')).toBeInTheDocument()
      }
    })
  })

  describe('Modal Behavior', () => {
    it('should close modal when close button is clicked', () => {
      render(<AuthModal onClose={mockOnClose} />)

      const closeButton = screen.getByRole('button', { name: '×' })
      fireEvent.click(closeButton)

      expect(mockOnClose).toHaveBeenCalled()
    })

    it('should close modal when clicking overlay', () => {
      render(<AuthModal onClose={mockOnClose} />)

      const overlay = screen.getByText('Welcome Back!').closest('.modal-overlay')
      if (overlay) {
        fireEvent.click(overlay)
      }

      expect(mockOnClose).toHaveBeenCalled()
    })

    it('should disable submit button while submitting', async () => {
      mockLogin.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      )

      render(<AuthModal onClose={mockOnClose} />)

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'test@example.com' },
      })
      fireEvent.change(screen.getByLabelText(/password/i), {
        target: { value: 'password123' },
      })

      const submitButton = screen.getByRole('button', { name: /login/i })
      fireEvent.click(submitButton)

      expect(submitButton).toBeDisabled()
      expect(screen.getByText(/please wait\.\.\./i)).toBeInTheDocument()

      await waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled()
      })
    })
  })

  describe('Form Validation', () => {
    it('should require email field', () => {
      render(<AuthModal onClose={mockOnClose} />)

      const emailInput = screen.getByLabelText(/email/i)
      expect(emailInput).toBeRequired()
      expect(emailInput).toHaveAttribute('type', 'email')
    })

    it('should require password field with minimum length', () => {
      render(<AuthModal onClose={mockOnClose} />)

      const passwordInput = screen.getByLabelText(/password/i)
      expect(passwordInput).toBeRequired()
      expect(passwordInput).toHaveAttribute('type', 'password')
      expect(passwordInput).toHaveAttribute('minlength', '8')
    })

    it('should require name field in register mode', () => {
      render(<AuthModal onClose={mockOnClose} initialMode="register" />)

      const nameInput = screen.getByLabelText(/name/i)
      expect(nameInput).toBeRequired()
    })
  })
})
