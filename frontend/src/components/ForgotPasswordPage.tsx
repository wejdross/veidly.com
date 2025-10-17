import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'
import './ForgotPasswordPage.css'

function ForgotPasswordPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await api.forgotPassword(email)
      setSuccess(true)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to send reset email')
    } finally {
      setLoading(false)
    }
  }

  if (success) {
    return (
      <div className="forgot-password-page">
        <div className="forgot-password-container">
          <div className="success-icon">✓</div>
          <h1>Check Your Email</h1>
          <p>
            If an account exists with the email <strong>{email}</strong>, you will receive
            a password reset link shortly.
          </p>
          <p className="hint">Please check your spam folder if you don't see the email.</p>
          <button onClick={() => navigate('/')} className="home-button">
            Go to Home
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="forgot-password-page">
      <div className="forgot-password-container">
        <h1>Reset Password</h1>
        <p className="subtitle">
          Enter your email address and we'll send you a link to reset your password.
        </p>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email Address</label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="your@email.com"
              required
              autoFocus
            />
          </div>

          {error && <div className="error-message">{error}</div>}

          <button type="submit" className="submit-button" disabled={loading}>
            {loading ? 'Sending...' : 'Send Reset Link'}
          </button>

          <button
            type="button"
            onClick={() => navigate('/')}
            className="cancel-button"
          >
            Cancel
          </button>
        </form>
      </div>
    </div>
  )
}

export default ForgotPasswordPage
