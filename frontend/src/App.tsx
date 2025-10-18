import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './AuthContext'
import LandingPage from './components/LandingPage'
import MapView from './components/MapView'
import AdminPanel from './components/AdminPanel'
import ProfilePage from './components/ProfilePage'
import PublicEventPage from './components/PublicEventPage'
import VerifyEmailPage from './components/VerifyEmailPage'
import ForgotPasswordPage from './components/ForgotPasswordPage'
import ResetPasswordPage from './components/ResetPasswordPage'
import VersionNotification from './components/VersionNotification'
import './App.css'

function ProtectedRoute({ children, adminOnly = false }: { children: React.ReactNode; adminOnly?: boolean }) {
  const { isAuthenticated, isAdmin } = useAuth()

  if (!isAuthenticated) {
    return <Navigate to="/map" replace />
  }

  if (adminOnly && !isAdmin) {
    return <Navigate to="/map" replace />
  }

  return <>{children}</>
}

function App() {
  return (
    <Router>
      <VersionNotification />
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/map" element={<MapView />} />
        <Route path="/event/:slug" element={<PublicEventPage />} />
        <Route path="/verify-email" element={<VerifyEmailPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="/reset-password" element={<ResetPasswordPage />} />
        <Route
          path="/profile"
          element={
            <ProtectedRoute>
              <ProfilePage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/profile/:userId"
          element={<ProfilePage />}
        />
        <Route
          path="/admin"
          element={
            <ProtectedRoute adminOnly>
              <AdminPanel />
            </ProtectedRoute>
          }
        />
      </Routes>
    </Router>
  )
}

export default App
