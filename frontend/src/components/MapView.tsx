import { useState, useEffect } from 'react'
import { MapContainer, TileLayer, Marker, Popup, useMapEvents, useMap } from 'react-leaflet'
import { Icon } from 'leaflet'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { api } from '../api'
import { API_BASE_URL } from '../config'
import { Event } from '../types'
import { getLanguagesDisplay } from '../languages'
import { sanitizeText } from '../utils/sanitize'
import EventForm from './EventForm'
import AuthModal from './AuthModal'
import EventsSidebar from './EventsSidebar'
import SearchPanel from './SearchPanel'
import ParticipantModal from './ParticipantModal'
import EmailVerificationBanner from './EmailVerificationBanner'
import ToastContainer, { showToast } from './ToastContainer'
import './MapView.css'

// Fix for default marker icon in React-Leaflet
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'

const defaultIcon = new Icon({
  iconUrl: markerIcon,
  iconRetinaUrl: markerIcon2x,
  shadowUrl: markerShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
})

function LocationMarker({ onLocationSelect }: { onLocationSelect: (lat: number, lng: number) => void }) {
  useMapEvents({
    click(e) {
      onLocationSelect(e.latlng.lat, e.latlng.lng)
    },
  })
  return null
}

function MapBoundsTracker({ onBoundsChange }: { onBoundsChange: (bounds: any) => void }) {
  const map = useMap()

  useEffect(() => {
    const updateBounds = () => {
      onBoundsChange(map.getBounds())
    }

    map.on('moveend', updateBounds)
    map.on('zoomend', updateBounds)

    // Initial bounds
    updateBounds()

    return () => {
      map.off('moveend', updateBounds)
      map.off('zoomend', updateBounds)
    }
  }, [map, onBoundsChange])

  return null
}

function MapView() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const { user, isAuthenticated, isAdmin, logout } = useAuth()
  const [events, setEvents] = useState<Event[]>([])
  const [visibleEvents, setVisibleEvents] = useState<Event[]>([])
  const [showForm, setShowForm] = useState(false)
  const [showAuth, setShowAuth] = useState(false)
  const [showSidebar, setShowSidebar] = useState(true)
  const [selectedLocation, setSelectedLocation] = useState<{ lat: number; lng: number } | null>(null)
  const [editingEvent, setEditingEvent] = useState<Event | null>(null)
  const [center, setCenter] = useState<[number, number]>([50.5, 10.0]) // Center of Europe (Germany) as default
  const [mapBounds, setMapBounds] = useState<any>(null)
  const [showParticipants, setShowParticipants] = useState(false)
  const [selectedEventForParticipants, setSelectedEventForParticipants] = useState<Event | null>(null)

  // Parse URL filters into state
  const getInitialFilters = () => {
    const filters: any = {}
    searchParams.forEach((value, key) => {
      filters[key] = value
    })
    return filters
  }

  // Load filters from URL on mount
  useEffect(() => {
    const filters = getInitialFilters()
    if (Object.keys(filters).length > 0) {
      loadEvents(filters)
    } else {
      loadEvents()
    }
  }, []) // Only run on mount

  // Handle redirect after successful authentication
  useEffect(() => {
    const returnTo = searchParams.get('returnTo')
    if (returnTo && isAuthenticated) {
      // Clear the returnTo parameter and navigate
      setSearchParams({})
      navigate(returnTo)
    }
  }, [isAuthenticated, searchParams, navigate, setSearchParams])

  useEffect(() => {

    // Try to get user's location
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          setCenter([position.coords.latitude, position.coords.longitude])
        },
        () => {
          console.log('Location access denied, using default location')
        }
      )
    }
  }, [])

  // Filter visible events when map bounds or events change
  useEffect(() => {
    if (!mapBounds) {
      setVisibleEvents(events)
      return
    }

    const filtered = events.filter((event) => {
      return mapBounds.contains([event.latitude, event.longitude])
    })
    setVisibleEvents(filtered)
  }, [mapBounds, events])

  const loadEvents = async (filters: any = {}) => {
    try {
      const data = await api.getEvents(filters)
      setEvents(data || [])
    } catch (error) {
      console.error('Failed to load events:', error)
      setEvents([])
    }
  }

  const handleSearch = (filters: any) => {
    // Update URL with filter params
    const params = new URLSearchParams()
    Object.entries(filters).forEach(([key, value]) => {
      if (value && value !== '' && value !== 'any') {
        params.set(key, value as string)
      }
    })
    setSearchParams(params)
    loadEvents(filters)
  }

  const handleClearSearch = () => {
    setSearchParams(new URLSearchParams()) // Clear URL params
    loadEvents()
  }

  // Disabled to prevent infinite render loop
  // Map center tracking in URL was causing excessive re-renders
  // const handleMapCenterChange = (center: { lat: number, lng: number }) => {
  //   setSearchParams(prev => {
  //     const newParams = new URLSearchParams(prev)
  //     newParams.set('lat', center.lat.toFixed(6))
  //     newParams.set('lng', center.lng.toFixed(6))
  //     return newParams
  //   })
  // }

  const handleLocationSelect = (lat: number, lng: number) => {
    if (!isAuthenticated) {
      setShowAuth(true)
      return
    }
    setSelectedLocation({ lat, lng })
    setShowForm(true)
  }

  const handleCreateEventClick = () => {
    if (!isAuthenticated) {
      setShowAuth(true)
      return
    }
    setShowForm(true)
  }

  const handleFormClose = () => {
    setShowForm(false)
    setSelectedLocation(null)
    setEditingEvent(null)
    loadEvents()
  }

  const handleEventClick = (event: Event) => {
    setCenter([event.latitude, event.longitude])
    setShowSidebar(false)
  }

  const handleEditEvent = (event: Event) => {
    setEditingEvent(event)
    setSelectedLocation({ lat: event.latitude, lng: event.longitude })
    setShowForm(true)
  }

  const handleJoinEvent = async (eventId: number) => {
    if (!isAuthenticated) {
      setShowAuth(true)
      return
    }

    try {
      await api.joinEvent(eventId)
      showToast('Successfully joined event!', 'success')
      loadEvents() // Reload to update participant count and is_participant status
    } catch (error: any) {
      showToast(error.response?.data?.error || 'Failed to join event', 'error')
    }
  }

  const handleLeaveEvent = async (eventId: number) => {
    if (!isAuthenticated) {
      return
    }

    try {
      await api.leaveEvent(eventId)
      showToast('Successfully left event!', 'success')
      loadEvents() // Reload to update participant count and is_participant status
    } catch (error: any) {
      showToast(error.response?.data?.error || 'Failed to leave event', 'error')
    }
  }

  const formatDateTime = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className="map-view">
      <EmailVerificationBanner />
      <nav className="map-navbar">
        <div className="logo" onClick={() => navigate('/')}>Veidly</div>
        <div className="nav-actions">
          {isAuthenticated ? (
            <>
              <button className="nav-button" onClick={handleCreateEventClick}>
                Create Event
              </button>
              {isAdmin && (
                <button className="nav-button admin-button" onClick={() => navigate('/admin')}>
                  Admin Panel
                </button>
              )}
              <div className="user-menu">
                <span className="user-name">Hi, {user?.name}!</span>
                <button className="nav-button" onClick={() => navigate('/profile')}>
                  My Profile
                </button>
                <button className="nav-button secondary" onClick={logout}>
                  Logout
                </button>
              </div>
            </>
          ) : (
            <button className="nav-button" onClick={() => setShowAuth(true)}>
              Login / Sign Up
            </button>
          )}
        </div>
      </nav>

      <div className="map-container-wrapper">
        <MapContainer
          center={center}
          zoom={5}
          className="map-container"
          style={{ height: '100%', width: '100%' }}
        >
          <TileLayer
            attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          />

          <LocationMarker onLocationSelect={handleLocationSelect} />
          <MapBoundsTracker onBoundsChange={setMapBounds} />

          {events.map((event) => (
            <Marker
              key={event.id}
              position={[event.latitude, event.longitude]}
              icon={defaultIcon}
            >
              <Popup>
                <div className="popup-content">
                  <h3
                    onClick={(e) => {
                      e.stopPropagation()
                      if (event.slug) navigate(`/event/${event.slug}`)
                    }}
                    style={{ cursor: event.slug ? 'pointer' : 'default', color: event.slug ? '#667eea' : 'inherit' }}
                    dangerouslySetInnerHTML={{ __html: sanitizeText(event.title) }}
                  />
                  <p className="popup-description" dangerouslySetInnerHTML={{ __html: sanitizeText(event.description) }} />
                  <div className="popup-details">
                    <p><strong>📅 When:</strong> {formatDateTime(event.start_time)}</p>
                    <p>
                      <strong>👤 Host:</strong>{' '}
                      <span
                        className="creator-link"
                        onClick={(e) => {
                          e.stopPropagation()
                          if (event.user_id) navigate(`/profile/${event.user_id}`)
                        }}
                        style={{ cursor: 'pointer', color: '#667eea', textDecoration: 'underline' }}
                        dangerouslySetInnerHTML={{ __html: sanitizeText(event.creator_name) }}
                      />
                    </p>
                    {event.max_participants && (
                      <p><strong>👥 Max participants:</strong> {event.max_participants}</p>
                    )}
                    {event.participant_count !== undefined && (
                      <p>
                        <strong>👥 Joined:</strong>{' '}
                        <span
                          className="participant-count-link"
                          onClick={(e) => {
                            e.stopPropagation()
                            setSelectedEventForParticipants(event)
                            setShowParticipants(true)
                          }}
                          style={{ cursor: 'pointer', color: '#667eea', textDecoration: 'underline', fontWeight: 600 }}
                        >
                          {event.participant_count}{event.max_participants ? `/${event.max_participants}` : ''}
                        </span>
                      </p>
                    )}
                    {event.event_languages && (
                      <p><strong>🗣️ Event Languages:</strong> {getLanguagesDisplay(event.event_languages)}</p>
                    )}
                    {event.creator_languages && (
                      <p><strong>🌍 Creator Languages:</strong> {event.creator_languages.split(',').map(code => {
                        const langNames: { [key: string]: string } = {
                          'de': 'German', 'fr': 'French', 'it': 'Italian', 'rm': 'Romansh',
                          'en': 'English', 'es': 'Spanish', 'pt': 'Portuguese', 'pl': 'Polish',
                          'tr': 'Turkish', 'ar': 'Arabic'
                        }
                        return langNames[code] || code
                      }).join(', ')}</p>
                    )}
                  </div>
                  <div className="popup-actions">
                    {isAuthenticated && user?.id === event.user_id && (
                      <button
                        className="edit-event-button"
                        onClick={(e) => {
                          e.stopPropagation()
                          handleEditEvent(event)
                        }}
                      >
                        ✏️ Edit Event
                      </button>
                    )}
                    {isAuthenticated && user?.id !== event.user_id && (
                      event.is_participant ? (
                        <button
                          className="leave-event-button"
                          onClick={(e) => {
                            e.stopPropagation()
                            if (event.id) handleLeaveEvent(event.id)
                          }}
                        >
                          ➖ Leave Event
                        </button>
                      ) : (
                        <button
                          className="join-event-button"
                          onClick={(e) => {
                            e.stopPropagation()
                            if (event.id) handleJoinEvent(event.id)
                          }}
                          disabled={event.max_participants !== undefined && event.participant_count !== undefined && event.participant_count >= event.max_participants}
                        >
                          {event.max_participants && event.participant_count !== undefined && event.participant_count >= event.max_participants ? '✓ Event Full' : '➕ Join Event'}
                        </button>
                      )
                    )}
                    {event.slug && (
                      <>
                        <button
                          className="share-event-button"
                          onClick={(e) => {
                            e.stopPropagation()
                            const shareUrl = `${window.location.origin}/event/${event.slug}`
                            navigator.clipboard.writeText(shareUrl).then(() => {
                              showToast('Event link copied to clipboard! Share it with your friends.', 'success')
                            }).catch(() => {
                              showToast('Failed to copy link. Please copy manually.', 'error')
                            })
                          }}
                          title="Share this event"
                        >
                          🔗 Share
                        </button>
                        {(event.is_participant || user?.id === event.user_id) && (
                          <button
                            className="download-ics-button"
                            onClick={(e) => {
                              e.stopPropagation()
                              window.open(`${API_BASE_URL}/public/events/${event.slug}/ics`, '_blank')
                            }}
                            title="Download ICS calendar file"
                          >
                            📅 Add to Calendar
                          </button>
                        )}
                      </>
                    )}
                  </div>
                  <div className="popup-tags">
                    {event.gender_restriction !== 'any' && (
                      <span className="tag">{event.gender_restriction}</span>
                    )}
                    {event.age_min > 0 || event.age_max < 99 ? (
                      <span className="tag">Age: {event.age_min}-{event.age_max}</span>
                    ) : null}
                    {event.smoking_allowed && <span className="tag tag-allowed">🚬 Smoking OK</span>}
                    {event.alcohol_allowed && <span className="tag tag-allowed">🍺 Alcohol OK</span>}
                    {!event.smoking_allowed && <span className="tag tag-not-allowed">🚭 No smoking</span>}
                    {!event.alcohol_allowed && <span className="tag tag-not-allowed">🚫 No alcohol</span>}
                  </div>
                </div>
              </Popup>
            </Marker>
          ))}
        </MapContainer>
      </div>

      {showForm && (
        <EventForm
          initialLocation={selectedLocation}
          onClose={handleFormClose}
          event={editingEvent || undefined}
        />
      )}

      {showAuth && (
        <AuthModal
          onClose={() => setShowAuth(false)}
        />
      )}

      <EventsSidebar
        events={visibleEvents}
        isOpen={showSidebar}
        onClose={() => setShowSidebar(false)}
        onEventClick={handleEventClick}
      />

      <SearchPanel
        onSearch={handleSearch}
        onClear={handleClearSearch}
        initialFilters={getInitialFilters()}
      />

      {showParticipants && selectedEventForParticipants && (
        <ParticipantModal
          eventId={selectedEventForParticipants.id!}
          eventTitle={selectedEventForParticipants.title}
          onClose={() => setShowParticipants(false)}
        />
      )}

      <ToastContainer />
    </div>
  )
}

export default MapView
