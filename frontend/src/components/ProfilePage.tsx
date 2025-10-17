import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { User, CATEGORIES } from '../types'
import { LANGUAGES, getLanguageDisplay } from '../languages'
import axios from 'axios'
import { API_BASE_URL } from '../config'
import './ProfilePage.css'

interface ProfileEvent {
  id: number
  title: string
  slug: string
  start_time: string
  category: string
  latitude: number
  longitude: number
  is_creator?: boolean
}

function ProfilePage() {
  const navigate = useNavigate()
  const { userId } = useParams<{ userId?: string }>()
  const { user: currentUser, logout } = useAuth()
  const [viewedUser, setViewedUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [createdEvents, setCreatedEvents] = useState<ProfileEvent[]>([])
  const [joinedEvents, setJoinedEvents] = useState<ProfileEvent[]>([])
  const [pastEvents, setPastEvents] = useState<ProfileEvent[]>([])

  // Accordion state - all collapsed by default
  const [createdExpanded, setCreatedExpanded] = useState(false)
  const [joinedExpanded, setJoinedExpanded] = useState(false)
  const [pastExpanded, setPastExpanded] = useState(false)
  const INITIAL_VISIBLE = 5

  const user = userId && viewedUser ? viewedUser : currentUser
  const isOwnProfile = !userId || (currentUser && userId === String(currentUser.id))
  const [isEditing, setIsEditing] = useState(false)
  const [formData, setFormData] = useState({
    name: user?.name || '',
    bio: user?.bio || '',
        threema: user?.threema || '',
    languages: user?.languages || '',
  })
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Load profile data
  useEffect(() => {
    // Check if viewing own profile or someone else's
    const isViewingOwnProfile = !userId || (currentUser && userId === String(currentUser.id))

    if (isViewingOwnProfile) {
      // Load own profile with events
      if (!currentUser) {
        // Not logged in, redirect to map
        navigate('/map')
        return
      }

      setLoading(true)
      axios.get(`${API_BASE_URL}/profile`)
        .then((response) => {
          const data = (response as any).data || {}
          if (data.user) {
            setViewedUser(data.user)
          }
          if (Array.isArray(data.created_events)) {
            setCreatedEvents(data.created_events)
          }
          if (Array.isArray(data.joined_events)) {
            setJoinedEvents(data.joined_events)
          }
          if (Array.isArray(data.past_events)) {
            setPastEvents(data.past_events)
          }
          setLoading(false)
        })
        .catch((err) => {
          console.error('Error loading own profile:', err)
          setLoading(false)
          navigate('/map')
        })
    } else {
      // Load someone else's profile
      setLoading(true)
      axios.get(`${API_BASE_URL}/profile/${userId}`)
        .then(response => {
          const data = response.data
          if (data.user) {
            setViewedUser(data.user)
          }
          if (Array.isArray(data.created_events)) {
            setCreatedEvents(data.created_events)
          }
          setLoading(false)
        })
        .catch((err) => {
          console.error('Error loading user profile:', err)
          setLoading(false)
          navigate('/map')
        })
    }
  }, [userId, currentUser, navigate])

  useEffect(() => {
    if (user) {
      setFormData({
        name: user.name || '',
        bio: user.bio || '',
                threema: user.threema || '',
        languages: user.languages || '',
      })
    }
  }, [user])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')
    setIsSubmitting(true)

    try {
      const response = await axios.put(`${API_BASE_URL}/profile`, formData)
      setSuccess('Profile updated successfully!')
      setIsEditing(false)

      // Update localStorage with new user data
      localStorage.setItem('user', JSON.stringify(response.data))
      window.location.reload() // Reload to update context
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update profile')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData(prev => ({
      ...prev,
      [e.target.name]: e.target.value
    }))
  }

  const handleLanguageToggle = (langCode: string) => {
    const currentLanguages = formData.languages ? formData.languages.split(',') : []
    const updatedLanguages = currentLanguages.includes(langCode)
      ? currentLanguages.filter(code => code !== langCode)
      : [...currentLanguages, langCode]

    setFormData(prev => ({
      ...prev,
      languages: updatedLanguages.join(',')
    }))
  }

  const getSelectedLanguages = () => {
    return formData.languages ? formData.languages.split(',') : []
  }

  // Accordion helpers
  const getVisibleEvents = (events: ProfileEvent[], expanded: boolean) => {
    return expanded ? events : events.slice(0, INITIAL_VISIBLE)
  }

  // ICS file generation
  const generateICS = (event: ProfileEvent) => {
    const formatDate = (dateString: string) => {
      const date = new Date(dateString)
      return date.toISOString().replace(/[-:]/g, '').split('.')[0] + 'Z'
    }

    const endTime = new Date(new Date(event.start_time).getTime() + 2 * 60 * 60 * 1000) // Add 2 hours

    const icsContent = [
      'BEGIN:VCALENDAR',
      'VERSION:2.0',
      'PRODID:-//Veidly//Event Calendar//EN',
      'CALSCALE:GREGORIAN',
      'METHOD:PUBLISH',
      'BEGIN:VEVENT',
      `DTSTART:${formatDate(event.start_time)}`,
      `DTEND:${formatDate(endTime.toISOString())}`,
      `DTSTAMP:${formatDate(new Date().toISOString())}`,
      `UID:${event.id}@veidly.com`,
      `SUMMARY:${event.title}`,
      `DESCRIPTION:${CATEGORIES[event.category as keyof typeof CATEGORIES]} event on Veidly`,
      `LOCATION:${event.latitude},${event.longitude}`,
      'STATUS:CONFIRMED',
      'SEQUENCE:0',
      'END:VEVENT',
      'END:VCALENDAR'
    ].join('\r\n')

    return icsContent
  }

  const downloadICS = (event: ProfileEvent) => {
    const icsContent = generateICS(event)
    const blob = new Blob([icsContent], { type: 'text/calendar;charset=utf-8' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `${event.slug}.ics`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(link.href)
  }

  if (loading) {
    return (
      <div className="profile-page">
        <div className="profile-content">
          <h1>Loading...</h1>
        </div>
      </div>
    )
  }

  if (!user) {
    navigate('/map')
    return null
  }

  return (
    <div className="profile-page">
      <nav className="profile-navbar">
        <div className="logo" onClick={() => navigate('/')}>Veidly</div>
        <div className="nav-actions">
          <button className="nav-button" onClick={() => navigate('/map')}>
            Back to Map
          </button>
          {isOwnProfile && (
            <button className="nav-button secondary" onClick={logout}>
              Logout
            </button>
          )}
        </div>
      </nav>

      <div className="profile-content">
        <h1>{isOwnProfile ? 'My Profile' : `${user.name}'s Profile`}</h1>

        {error && <div className="error-message">{error}</div>}
        {success && <div className="success-message">{success}</div>}

        <div className="profile-card">
          {isOwnProfile && isEditing ? (
            <form onSubmit={handleSubmit} className="profile-form">
              <div className="form-group">
                <label htmlFor="bio">Bio (optional)</label>
                <textarea
                  id="bio"
                  name="bio"
                  value={formData.bio}
                  onChange={handleChange}
                  placeholder="Tell us about yourself..."
                  rows={4}
                  maxLength={1000}
                />
                <small className="form-hint">Max 1000 characters</small>
              </div>

              <div className="form-group">
                <label htmlFor="threema">Default contact method (optional)</label>
                <input
                  type="text"
                  id="threema"
                  name="threema"
                  value={formData.threema}
                  onChange={handleChange}
                  placeholder="@username or contact info"
                />
                <small className="form-hint">Will be used to auto-fill event contact info</small>
              </div>

              <div className="form-group">
                <label>Spoken Languages (optional)</label>
                <small className="form-hint" style={{display: 'block', marginBottom: '0.75rem'}}>
                  Select languages you speak - will be auto-filled when creating events
                </small>
                <div className="language-selector-grid">
                  {LANGUAGES.map(lang => (
                    <button
                      key={lang.code}
                      type="button"
                      className={`language-option ${getSelectedLanguages().includes(lang.code) ? 'selected' : ''}`}
                      onClick={() => handleLanguageToggle(lang.code)}
                    >
                      <span className="language-flag">{lang.flag}</span>
                      <span className="language-name">{lang.name}</span>
                    </button>
                  ))}
                </div>
                {formData.languages && (
                  <div className="selected-languages-count">
                    {formData.languages.split(',').length} language(s) selected
                  </div>
                )}
              </div>

              <div className="form-actions">
                <button type="button" className="button-secondary" onClick={() => setIsEditing(false)}>
                  Cancel
                </button>
                <button type="submit" className="button-primary" disabled={isSubmitting}>
                  {isSubmitting ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          ) : (
            <div className="profile-view">
              <div className="profile-header">
                <div className="profile-avatar">
                  {user.name.charAt(0).toUpperCase()}
                </div>
                <div>
                  <h2>{user.name}</h2>
                  <p className="profile-email">{user.email}</p>
                  {user.is_admin && <span className="admin-badge">Admin</span>}
                </div>
              </div>

              <div className="profile-details">
                {user.bio && (
                  <div className="detail-group">
                    <label>Bio</label>
                    <p>{user.bio}</p>
                  </div>
                )}

                <div className="detail-group">
                  <label>Default contact method</label>
                  <p>{user.threema || 'Not set'}</p>
                </div>

                <div className="detail-group">
                  <label>Languages</label>
                  <p>
                    {user.languages
                      ? user.languages.split(',').map(code => getLanguageDisplay(code.trim())).join(', ')
                      : 'Not set'}
                  </p>
                </div>

                <div className="detail-group">
                  <label>Member Since</label>
                  <p>{new Date(user.created_at).toLocaleDateString()}</p>
                </div>
              </div>

              {isOwnProfile && (
                <button className="button-primary" onClick={() => setIsEditing(true)}>
                  Edit Profile
                </button>
              )}
            </div>
          )}
        </div>

        {/* Events Sections */}
        <div className="events-section">
          {isOwnProfile && createdEvents.length > 0 && (
            <div className="event-category">
              <h2
                className="category-title accordion-header"
                onClick={() => setCreatedExpanded(!createdExpanded)}
              >
                <span className="category-icon">🎯</span>
                Events You're Hosting ({createdEvents.length})
                <span className="accordion-arrow">{createdExpanded ? '▼' : '▶'}</span>
              </h2>
              {createdExpanded && (
                <>
                  <div className="events-grid">
                    {getVisibleEvents(createdEvents, createdExpanded || createdEvents.length <= INITIAL_VISIBLE).map((event) => (
                      <div
                        key={event.id}
                        className="event-card"
                        onClick={() => navigate(`/event/${event.slug}`)}
                      >
                        <div className="event-category-badge">
                          {CATEGORIES[event.category as keyof typeof CATEGORIES]}
                        </div>
                        <h3 className="event-title">{event.title}</h3>
                        <div className="event-date">
                          {new Date(event.start_time).toLocaleString('en-US', {
                            month: 'short',
                            day: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit'
                          })}
                        </div>
                      </div>
                    ))}
                  </div>
                  {createdEvents.length > INITIAL_VISIBLE && (
                    <button
                      className="show-more-btn"
                      onClick={() => setCreatedExpanded(!createdExpanded)}
                    >
                      {createdExpanded ? '▲ Show Less' : `▼ Show All (${createdEvents.length})`}
                    </button>
                  )}
                </>
              )}
            </div>
          )}

          {isOwnProfile && joinedEvents.length > 0 && (
            <div className="event-category">
              <h2
                className="category-title accordion-header"
                onClick={() => setJoinedExpanded(!joinedExpanded)}
              >
                <span className="category-icon">🤝</span>
                Events You've Joined ({joinedEvents.length})
                <span className="accordion-arrow">{joinedExpanded ? '▼' : '▶'}</span>
              </h2>
              {joinedExpanded && (
                <>
                  <div className="events-grid">
                    {getVisibleEvents(joinedEvents, joinedExpanded || joinedEvents.length <= INITIAL_VISIBLE).map((event) => (
                      <div
                        key={event.id}
                        className="event-card event-card-with-actions"
                      >
                        <div onClick={() => navigate(`/event/${event.slug}`)} style={{ cursor: 'pointer' }}>
                          <div className="event-category-badge">
                            {CATEGORIES[event.category as keyof typeof CATEGORIES]}
                          </div>
                          <h3 className="event-title">{event.title}</h3>
                          <div className="event-date">
                            {new Date(event.start_time).toLocaleString('en-US', {
                              month: 'short',
                              day: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit'
                            })}
                          </div>
                        </div>
                        <button
                          className="ics-download-btn"
                          onClick={(e) => {
                            e.stopPropagation()
                            downloadICS(event)
                          }}
                          title="Download calendar invitation"
                        >
                          📅 Add to Calendar
                        </button>
                      </div>
                    ))}
                  </div>
                  {joinedEvents.length > INITIAL_VISIBLE && (
                    <button
                      className="show-more-btn"
                      onClick={() => setJoinedExpanded(!joinedExpanded)}
                    >
                      {joinedExpanded ? '▲ Show Less' : `▼ Show All (${joinedEvents.length})`}
                    </button>
                  )}
                </>
              )}
            </div>
          )}

          {isOwnProfile && pastEvents.length > 0 && (
            <div className="event-category">
              <h2
                className="category-title accordion-header"
                onClick={() => setPastExpanded(!pastExpanded)}
              >
                <span className="category-icon">📅</span>
                Past Events ({pastEvents.length})
                <span className="accordion-arrow">{pastExpanded ? '▼' : '▶'}</span>
              </h2>
              {pastExpanded && (
                <>
                  <div className="events-grid">
                    {getVisibleEvents(pastEvents, pastExpanded || pastEvents.length <= INITIAL_VISIBLE).map((event) => (
                      <div
                        key={event.id}
                        className="event-card past-event"
                        onClick={() => navigate(`/event/${event.slug}`)}
                      >
                        <div className="event-category-badge">
                          {CATEGORIES[event.category as keyof typeof CATEGORIES]}
                        </div>
                        <h3 className="event-title">{event.title}</h3>
                        <div className="event-date">
                          {new Date(event.start_time).toLocaleString('en-US', {
                            month: 'short',
                            day: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit'
                          })}
                        </div>
                        <div className="event-role">
                          {event.is_creator ? '👑 Hosted' : '✅ Attended'}
                        </div>
                      </div>
                    ))}
                  </div>
                  {pastEvents.length > INITIAL_VISIBLE && (
                    <button
                      className="show-more-btn"
                      onClick={() => setPastExpanded(!pastExpanded)}
                    >
                      {pastExpanded ? '▲ Show Less' : `▼ Show All (${pastEvents.length})`}
                    </button>
                  )}
                </>
              )}
            </div>
          )}

          {!isOwnProfile && createdEvents.length > 0 && (
            <div className="event-category">
              <h2
                className="category-title accordion-header"
                onClick={() => setCreatedExpanded(!createdExpanded)}
              >
                <span className="category-icon">🎯</span>
                {user?.name}'s Upcoming Events ({createdEvents.length})
                <span className="accordion-arrow">{createdExpanded ? '▼' : '▶'}</span>
              </h2>
              {createdExpanded && (
                <>
                  <div className="events-grid">
                    {getVisibleEvents(createdEvents, createdExpanded || createdEvents.length <= INITIAL_VISIBLE).map((event) => (
                      <div
                        key={event.id}
                        className="event-card"
                        onClick={() => navigate(`/event/${event.slug}`)}
                      >
                        <div className="event-category-badge">
                          {CATEGORIES[event.category as keyof typeof CATEGORIES]}
                        </div>
                        <h3 className="event-title">{event.title}</h3>
                        <div className="event-date">
                          {new Date(event.start_time).toLocaleString('en-US', {
                            month: 'short',
                            day: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit'
                          })}
                        </div>
                      </div>
                    ))}
                  </div>
                  {createdEvents.length > INITIAL_VISIBLE && (
                    <button
                      className="show-more-btn"
                      onClick={() => setCreatedExpanded(!createdExpanded)}
                    >
                      {createdExpanded ? '▲ Show Less' : `▼ Show All (${createdEvents.length})`}
                    </button>
                  )}
                </>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default ProfilePage
