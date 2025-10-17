import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import axios from 'axios'
import { User, Event, CATEGORIES } from '../types'
import ToastContainer, { showToast } from './ToastContainer'
import { API_BASE_URL } from '../config'
import './AdminPanel.css'

function AdminPanel() {
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const [users, setUsers] = useState<User[]>([])
  const [events, setEvents] = useState<Event[]>([])
  const [activeTab, setActiveTab] = useState<'users' | 'events'>('users')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadData()
  }, [activeTab])

  const loadData = async () => {
    setLoading(true)
    try {
      if (activeTab === 'users') {
        const response = await axios.get<User[]>(`${API_BASE_URL}/admin/users`)
        setUsers(response.data)
      } else {
        const response = await axios.get<Event[]>(`${API_BASE_URL}/admin/events`)
        setEvents(response.data)
      }
    } catch (error) {
      console.error('Failed to load data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleBlockUser = async (userId: number) => {
    try {
      await axios.put(`${API_BASE_URL}/admin/users/${userId}/block`)
      showToast('User blocked successfully', 'success')
      loadData()
    } catch (error) {
      showToast('Failed to block user', 'error')
    }
  }

  const handleUnblockUser = async (userId: number) => {
    try {
      await axios.put(`${API_BASE_URL}/admin/users/${userId}/unblock`)
      showToast('User unblocked successfully', 'success')
      loadData()
    } catch (error) {
      showToast('Failed to unblock user', 'error')
    }
  }

  const handleVerifyUserEmail = async (userId: number) => {
    try {
      await axios.put(`${API_BASE_URL}/admin/users/${userId}/verify-email`)
      showToast('User email verified successfully', 'success')
      loadData()
    } catch (error) {
      showToast('Failed to verify user email', 'error')
    }
  }

  const handleDeleteEvent = async (eventId: number) => {
    if (!confirm('Are you sure you want to delete this event?')) return

    try {
      await axios.delete(`${API_BASE_URL}/admin/events/${eventId}`)
      showToast('Event deleted successfully', 'success')
      loadData()
    } catch (error) {
      showToast('Failed to delete event', 'error')
    }
  }

  return (
    <div className="admin-panel">
      <nav className="admin-navbar">
        <div className="logo" onClick={() => navigate('/')}>Veidly Admin</div>
        <div className="admin-nav-actions">
          <button className="nav-button" onClick={() => navigate('/map')}>
            Back to Map
          </button>
          <span className="admin-user-name">{user?.name}</span>
          <button className="nav-button secondary" onClick={logout}>
            Logout
          </button>
        </div>
      </nav>

      <div className="admin-content">
        <h1>Admin Panel</h1>

        <div className="admin-tabs">
          <button
            className={`tab ${activeTab === 'users' ? 'active' : ''}`}
            onClick={() => setActiveTab('users')}
          >
            Users Management
          </button>
          <button
            className={`tab ${activeTab === 'events' ? 'active' : ''}`}
            onClick={() => setActiveTab('events')}
          >
            Events Management
          </button>
        </div>

        {loading ? (
          <div className="loading">Loading...</div>
        ) : (
          <>
            {activeTab === 'users' ? (
              <div className="admin-table-container">
                <h2>All Users ({users?.length || 0})</h2>
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>Name</th>
                      <th>Email</th>
                      <th>Email Verified</th>
                      <th>Admin</th>
                      <th>Status</th>
                      <th>Created</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((u) => (
                      <tr key={u.id} className={u.is_blocked ? 'blocked-row' : ''}>
                        <td>{u.id}</td>
                        <td>{u.name}</td>
                        <td>{u.email}</td>
                        <td>
                          <span className={`status-badge ${u.email_verified ? 'verified' : 'unverified'}`}>
                            {u.email_verified ? '✅ Verified' : '⚠️ Unverified'}
                          </span>
                        </td>
                        <td>{u.is_admin ? '✅ Admin' : '-'}</td>
                        <td>
                          <span className={`status-badge ${u.is_blocked ? 'blocked' : 'active'}`}>
                            {u.is_blocked ? 'Blocked' : 'Active'}
                          </span>
                        </td>
                        <td>{new Date(u.created_at).toLocaleDateString()}</td>
                        <td>
                          <div className="action-buttons">
                            {!u.email_verified && (
                              <button
                                className="action-button verify"
                                onClick={() => handleVerifyUserEmail(u.id)}
                                title="Manually verify user email"
                              >
                                Verify Email
                              </button>
                            )}
                            {!u.is_admin && (
                              u.is_blocked ? (
                                <button
                                  className="action-button unblock"
                                  onClick={() => handleUnblockUser(u.id)}
                                >
                                  Unblock
                                </button>
                              ) : (
                                <button
                                  className="action-button block"
                                  onClick={() => handleBlockUser(u.id)}
                                >
                                  Block
                                </button>
                              )
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="admin-table-container">
                <h2>All Events ({events?.length || 0})</h2>
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>Title</th>
                      <th>Category</th>
                      <th>Creator</th>
                      <th>Date</th>
                      <th>Location</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {events && events.length > 0 ? events.map((e) => (
                      <tr key={e.id}>
                        <td>{e.id}</td>
                        <td>{e.title}</td>
                        <td>{CATEGORIES[e.category as keyof typeof CATEGORIES]}</td>
                        <td>
                          {e.creator_name}<br/>
                          <small>{e.user_email}</small>
                        </td>
                        <td>{new Date(e.start_time).toLocaleString()}</td>
                        <td>
                          {e.latitude.toFixed(4)}, {e.longitude.toFixed(4)}
                        </td>
                        <td>
                          <button
                            className="action-button delete"
                            onClick={() => handleDeleteEvent(e.id!)}
                          >
                            Delete
                          </button>
                        </td>
                      </tr>
                    )) : (
                      <tr>
                        <td colSpan={7} style={{ textAlign: 'center', padding: '2rem' }}>
                          No events found
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            )}
          </>
        )}
      </div>

      <ToastContainer />
    </div>
  )
}

export default AdminPanel
