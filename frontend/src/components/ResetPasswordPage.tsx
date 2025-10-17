import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../api'
import './ResetPasswordPage.css'

function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState('')

  const token = searchParams.get('token')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!token) {
      setError('Invalid reset link')
      return
    }

    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters long')
      return
    }

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    setLoading(true)

    try {
      await api.resetPassword(token, newPassword)
      setSuccess(true)

      // Redirect to home page after 3 seconds
      setTimeout(() => {
        navigate('/')
      }, 3000)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to reset password')
    } finally {
      setLoading(false)
    }
  }

  if (!token) {
    return (
      <div className="reset-password-page">
        <div className="reset-password-container">
          <div className="error-icon">✕</div>
          <h1>Invalid Link</h1>
          <p>This password reset link is invalid or has expired.</p>
          <button onClick={() => navigate('/forgot-password')} className="action-button">
            Request New Link
          </button>
          <button onClick={() => navigate('/')} className="cancel-button">
            Go to Home
          </button>
        </div>
      </div>
    )
  }

  if (success) {
    return (
      <div className="reset-password-page">
        <div className="reset-password-container">
          <div className="success-icon">✓</div>
          <h1>Password Reset Successful!</h1>
          <p>Your password has been reset successfully.</p>
          <p className="hint">You can now log in with your new password.</p>
          <p className="redirect-message">Redirecting you to the home page...</p>
          <button onClick={() => navigate('/')} className="action-button">
            Go to Home
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="reset-password-page">
      <div className="reset-password-container">
        <h1>Create New Password</h1>
        <p className="subtitle">
          Enter your new password below.
        </p>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="newPassword">New Password</label>
            <input
              id="newPassword"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="At least 8 characters"
              required
              autoFocus
              minLength={8}
            />
          </div>

          <div className="form-group">
            <label htmlFor="confirmPassword">Confirm Password</label>
            <input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Re-enter your password"
              required
              minLength={8}
            />
          </div>

          {error && <div className="error-message">{error}</div>}

          <button type="submit" className="submit-button" disabled={loading}>
            {loading ? 'Resetting...' : 'Reset Password'}
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

export default ResetPasswordPage
