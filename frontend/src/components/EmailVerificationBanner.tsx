import { useState } from 'react'
import { useAuth } from '../AuthContext'
import { api } from '../api'
import './EmailVerificationBanner.css'

function EmailVerificationBanner() {
  const { user } = useAuth()
  const [resending, setResending] = useState(false)
  const [message, setMessage] = useState('')
  const [dismissed, setDismissed] = useState(false)

  // Don't show banner if user is verified, is admin, dismissed, or not logged in
  // Admins can manually verify their email via admin panel
  if (!user || user.email_verified || user.is_admin || dismissed) {
    return null
  }

  const handleResend = async () => {
    setResending(true)
    setMessage('')

    try {
      await api.resendVerificationEmail(user.email)
      setMessage('Verification email sent! Please check your inbox.')
    } catch (error: any) {
      setMessage(error.response?.data?.error || 'Failed to send verification email')
    } finally {
      setResending(false)
    }
  }

  return (
    <div className="email-verification-banner">
      <div className="banner-content">
        <div className="banner-icon">⚠️</div>
        <div className="banner-text">
          <strong>Please verify your email address</strong>
          <p>
            We sent a verification link to <strong>{user.email}</strong>.
            You need to verify your email before creating events.
          </p>
          {message && <p className="banner-message">{message}</p>}
        </div>
        <div className="banner-actions">
          <button
            onClick={handleResend}
            className="resend-button"
            disabled={resending}
          >
            {resending ? 'Sending...' : 'Resend Email'}
          </button>
          <button
            onClick={() => setDismissed(true)}
            className="dismiss-button"
            title="Dismiss"
          >
            ✕
          </button>
        </div>
      </div>
    </div>
  )
}

export default EmailVerificationBanner
