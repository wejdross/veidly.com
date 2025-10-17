import { useState, FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import './AuthModal.css'

interface AuthModalProps {
  onClose: () => void
  initialMode?: 'login' | 'register'
}

function AuthModal({ onClose, initialMode = 'login' }: AuthModalProps) {
  const navigate = useNavigate()
  const { login, register } = useAuth()
  const [mode, setMode] = useState<'login' | 'register'>(initialMode)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleForgotPassword = () => {
    onClose()
    navigate('/forgot-password')
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setIsSubmitting(true)

    try {
      if (mode === 'login') {
        await login(email, password)
      } else {
        if (!name.trim()) {
          setError('Name is required')
          setIsSubmitting(false)
          return
        }
        await register(email, password, name)
      }
      onClose()
    } catch (err: any) {
      setError(err.response?.data?.error || `${mode === 'login' ? 'Login' : 'Registration'} failed`)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content auth-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>{mode === 'login' ? 'Welcome Back!' : 'Join Veidly'}</h2>
          <button className="close-button" onClick={onClose}>&times;</button>
        </div>

        <form onSubmit={handleSubmit} className="auth-form">
          {error && <div className="error-message">{error}</div>}

          {mode === 'register' && (
            <div className="form-group">
              <label htmlFor="name">Name</label>
              <input
                type="text"
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                placeholder="Your name"
              />
            </div>
          )}

          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder="you@example.com"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              placeholder="Minimum 8 characters"
            />
            {mode === 'login' && (
              <button
                type="button"
                onClick={handleForgotPassword}
                className="forgot-password-link"
              >
                Forgot password?
              </button>
            )}
          </div>

          <button type="submit" className="button-primary" disabled={isSubmitting}>
            {isSubmitting ? 'Please wait...' : mode === 'login' ? 'Login' : 'Create Account'}
          </button>

          <div className="auth-switch">
            {mode === 'login' ? (
              <p>
                Don't have an account?{' '}
                <button type="button" onClick={() => setMode('register')} className="link-button">
                  Sign up
                </button>
              </p>
            ) : (
              <p>
                Already have an account?{' '}
                <button type="button" onClick={() => setMode('login')} className="link-button">
                  Login
                </button>
              </p>
            )}
          </div>
        </form>
      </div>
    </div>
  )
}

export default AuthModal
