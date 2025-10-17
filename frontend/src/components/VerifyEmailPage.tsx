import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../api'
import './VerifyEmailPage.css'

function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [status, setStatus] = useState<'verifying' | 'success' | 'error'>('verifying')
  const [message, setMessage] = useState('')

  useEffect(() => {
    const token = searchParams.get('token')

    if (!token) {
      setStatus('error')
      setMessage('Invalid verification link')
      return
    }

    const verifyEmail = async () => {
      try {
        await api.verifyEmail(token)
        setStatus('success')
        setMessage('Your email has been verified successfully!')

        // Redirect to home page after 3 seconds
        setTimeout(() => {
          navigate('/')
        }, 3000)
      } catch (error: any) {
        setStatus('error')
        setMessage(error.response?.data?.error || 'Email verification failed')
      }
    }

    verifyEmail()
  }, [searchParams, navigate])

  return (
    <div className="verify-email-page">
      <div className="verify-email-container">
        {status === 'verifying' && (
          <>
            <div className="spinner"></div>
            <h1>Verifying your email...</h1>
            <p>Please wait while we verify your email address</p>
          </>
        )}

        {status === 'success' && (
          <>
            <div className="success-icon">✓</div>
            <h1>Email Verified!</h1>
            <p>{message}</p>
            <p className="redirect-message">Redirecting you to the home page...</p>
            <button onClick={() => navigate('/')} className="home-button">
              Go to Home
            </button>
          </>
        )}

        {status === 'error' && (
          <>
            <div className="error-icon">✕</div>
            <h1>Verification Failed</h1>
            <p>{message}</p>
            <button onClick={() => navigate('/')} className="home-button">
              Go to Home
            </button>
          </>
        )}
      </div>
    </div>
  )
}

export default VerifyEmailPage
