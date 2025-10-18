import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet'
import { Icon } from 'leaflet'
import { Event, CATEGORIES } from '../types'
import { useAuth } from '../AuthContext'
import { API_BASE_URL } from '../config'
import { api } from '../api'
import { getLanguagesDisplay } from '../languages'
import { sanitizeText } from '../utils/sanitize'
import ToastContainer, { showToast } from './ToastContainer'
import EventComments from './EventComments'
import 'leaflet/dist/leaflet.css'

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

export default function PublicEventPage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { isAuthenticated, user } = useAuth()
  const [event, setEvent] = useState<Event | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchEvent = async () => {
      try {
        setLoading(true)
        const data = await api.getPublicEvent(slug!)
        console.log('Fetched event data:', {
          id: data.id,
          is_participant: data.is_participant,
          user_id: data.user_id
        })
        setEvent(data)
      } catch (err) {
        setError('Failed to load event')
      } finally {
        setLoading(false)
      }
    }

    if (slug) {
      fetchEvent()
    }
  }, [slug])

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      weekday: 'short',
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const handleJoinEvent = async () => {
    if (!event?.id || !slug) return

    try {
      await api.joinEvent(event.id)
      showToast('Successfully joined event!', 'success')
      // Refetch event to update is_participant status
      const data = await api.getPublicEvent(slug)
      console.log('After joining - refetched event:', {
        id: data.id,
        is_participant: data.is_participant,
        user_id: data.user_id
      })
      setEvent(data)
    } catch (error: any) {
      showToast(error.response?.data?.error || 'Failed to join event', 'error')
    }
  }

  const handleLeaveEvent = async () => {
    if (!event?.id || !slug) return

    try {
      await api.leaveEvent(event.id)
      showToast('Successfully left event!', 'success')
      // Refetch event to update is_participant status
      const data = await api.getPublicEvent(slug)
      setEvent(data)
    } catch (error: any) {
      showToast(error.response?.data?.error || 'Failed to leave event', 'error')
    }
  }

  const handleDownloadICS = () => {
    window.open(`${API_BASE_URL}/public/events/${slug}/ics`, '_blank')
  }

  if (loading) {
    return (
      <div className="public-event-page">
        <div className="loading">Loading event...</div>
      </div>
    )
  }

  if (error || !event) {
    return (
      <div className="public-event-page">
        <div className="error-message">
          <h2>Event Not Found</h2>
          <p>{error || 'This event does not exist or has been removed.'}</p>
          <button onClick={() => navigate('/map')} className="btn-primary">
            Browse Events
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="public-event-page">
      <header className="public-header">
        <div className="logo-container" onClick={() => navigate('/')} style={{ cursor: 'pointer' }}>
          <img src="/logo.svg" alt="Veidly" className="logo-icon" />
          <h1>Veidly</h1>
        </div>
        <div className="header-actions">
          {isAuthenticated ? (
            <button onClick={() => navigate('/map')} className="btn-secondary">
              Go to Map
            </button>
          ) : (
            <>
              <button onClick={() => navigate('/map')} className="btn-secondary">
                Browse Events
              </button>
              <button onClick={() => navigate(`/?returnTo=/event/${slug}`)} className="btn-primary">
                Sign In
              </button>
            </>
          )}
        </div>
      </header>

      <div className="event-details-container">
        <div className="event-header">
          <h1 dangerouslySetInnerHTML={{ __html: sanitizeText(event.title) }} />
          <span className="event-category">
            {CATEGORIES[event.category as keyof typeof CATEGORIES] || event.category}
          </span>
        </div>

        <div className="event-info-grid">
          <div className="event-info-section">
            <h3>About</h3>
            <p className="event-description" dangerouslySetInnerHTML={{ __html: sanitizeText(event.description) }} />

            <h3>When</h3>
            <p>
              <strong>Start:</strong> {formatDate(event.start_time)}
            </p>
            {event.end_time && (
              <p>
                <strong>End:</strong> {formatDate(event.end_time)}
              </p>
            )}

            <h3>Organizer</h3>
            <p>
              <span dangerouslySetInnerHTML={{ __html: sanitizeText(event.creator_name) }} />
            </p>

            {event.event_languages && (
              <>
                <h3>Event Languages</h3>
                <p>{getLanguagesDisplay(event.event_languages)}</p>
              </>
            )}

            {event.max_participants && (
              <>
                <h3>Participants</h3>
                <p>
                  {event.participant_count || 0} / {event.max_participants} joined
                </p>
              </>
            )}

            <h3>Details</h3>
            <ul className="event-details-list">
              {event.gender_restriction && event.gender_restriction !== 'any' && (
                <li>Gender: {event.gender_restriction}</li>
              )}
              {event.age_min > 0 || event.age_max < 99 ? (
                <li>
                  Age: {event.age_min} - {event.age_max}
                </li>
              ) : null}
              <li>Smoking: {event.smoking_allowed ? 'Allowed' : 'Not allowed'}</li>
              <li>Alcohol: {event.alcohol_allowed ? 'Allowed' : 'Not allowed'}</li>
            </ul>
          </div>

          <div className="event-map-section">
            <h3>Location</h3>
            <div className="map-container">
              <MapContainer
                center={[event.latitude, event.longitude]}
                zoom={13}
                style={{ height: '400px', width: '100%' }}
              >
                <TileLayer
                  url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                  attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                />
                <Marker position={[event.latitude, event.longitude]} icon={defaultIcon}>
                  <Popup>{event.title}</Popup>
                </Marker>
              </MapContainer>
            </div>
            <a
              href={`https://www.google.com/maps/search/?api=1&query=${event.latitude},${event.longitude}`}
              target="_blank"
              rel="noopener noreferrer"
              className="btn-maps"
            >
              📍 Open in Google Maps
            </a>
          </div>
        </div>

        {!isAuthenticated ? (
          <div className="join-cta">
            <p>Want to join this event?</p>
            <button onClick={() => navigate('/')} className="btn-primary btn-large">
              Sign In to Join
            </button>
          </div>
        ) : user?.id !== event.user_id && (
          <div className="action-buttons">
            {event.is_participant ? (
              <button onClick={handleLeaveEvent} className="btn-leave btn-large">
                ➖ Leave Event
              </button>
            ) : (
              <button
                onClick={handleJoinEvent}
                className="btn-join btn-large"
                disabled={event.max_participants !== undefined && event.participant_count !== undefined && event.participant_count >= event.max_participants}
              >
                {event.max_participants && event.participant_count !== undefined && event.participant_count >= event.max_participants ? '✓ Event Full' : '➕ Join Event'}
              </button>
            )}
            <button onClick={handleDownloadICS} className="btn-calendar btn-large">
              📅 Add to Calendar
            </button>
          </div>
        )}

        {isAuthenticated && user?.id === event.user_id && (
          <div className="action-buttons">
            <button onClick={handleDownloadICS} className="btn-calendar btn-large">
              📅 Add to Calendar
            </button>
          </div>
        )}

        {/* Comments Section */}
        {event.id && (
          <EventComments
            eventId={event.id}
            isParticipant={event.is_participant || user?.id === event.user_id || false}
          />
        )}

        <ToastContainer />
      </div>

      <style>{`
        .public-event-page {
          min-height: 100vh;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }

        .public-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 1rem 2rem;
          background: rgba(255, 255, 255, 0.1);
          backdrop-filter: blur(10px);
          color: white;
        }

        .public-header .logo-container {
          display: flex;
          align-items: center;
          gap: 1rem;
          cursor: pointer;
          transition: transform 0.3s ease;
        }

        .public-header .logo-container:hover {
          transform: scale(1.05);
        }

        .public-header .logo-icon {
          width: 40px;
          height: 40px;
          filter: drop-shadow(2px 2px 4px rgba(0, 0, 0, 0.2));
        }

        .public-header h1 {
          margin: 0;
          font-size: 1.8rem;
        }

        .header-actions {
          display: flex;
          gap: 1rem;
        }

        .event-details-container {
          max-width: 1200px;
          margin: 2rem auto;
          padding: 2rem;
          background: white;
          border-radius: 12px;
          box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
        }

        .event-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 2rem;
          padding-bottom: 1rem;
          border-bottom: 2px solid #f0f0f0;
        }

        .event-header h1 {
          margin: 0;
          color: #333;
        }

        .event-category {
          padding: 0.5rem 1rem;
          background: #667eea;
          color: white;
          border-radius: 20px;
          font-size: 0.9rem;
        }

        .event-info-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 2rem;
        }

        @media (max-width: 768px) {
          .event-info-grid {
            grid-template-columns: 1fr;
          }
        }

        .event-info-section h3,
        .event-map-section h3 {
          color: #667eea;
          margin-top: 1.5rem;
          margin-bottom: 0.5rem;
        }

        .event-description {
          line-height: 1.6;
          color: #555;
        }

        .event-details-list {
          list-style: none;
          padding: 0;
        }

        .event-details-list li {
          padding: 0.5rem 0;
          color: #555;
        }

        .join-cta {
          margin-top: 2rem;
          padding: 2rem;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
          text-align: center;
          border-radius: 12px;
        }

        .join-cta p {
          margin: 0 0 1rem 0;
          font-size: 1.2rem;
        }

        .btn-large {
          padding: 1rem 2rem;
          font-size: 1.1rem;
        }

        .loading,
        .error-message {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          min-height: 50vh;
          color: white;
          text-align: center;
        }

        .error-message h2 {
          margin-bottom: 1rem;
        }

        .btn-primary,
        .btn-secondary {
          padding: 0.75rem 1.5rem;
          border: none;
          border-radius: 8px;
          cursor: pointer;
          font-size: 1rem;
          transition: all 0.3s;
        }

        .btn-primary {
          background: white;
          color: #667eea;
        }

        .btn-primary:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
        }

        .btn-secondary {
          background: rgba(255, 255, 255, 0.2);
          color: white;
          border: 2px solid white;
        }

        .btn-secondary:hover {
          background: rgba(255, 255, 255, 0.3);
        }

        .action-buttons {
          margin-top: 2rem;
          display: flex;
          gap: 1rem;
          flex-wrap: wrap;
          justify-content: center;
        }

        .btn-join {
          background: linear-gradient(135deg, #34d399 0%, #10b981 100%);
          color: white;
          padding: 1rem 2rem;
          border: none;
          border-radius: 8px;
          cursor: pointer;
          font-size: 1.1rem;
          font-weight: 700;
          transition: all 0.3s;
          flex: 1;
          min-width: 200px;
        }

        .btn-join:hover:not(:disabled) {
          transform: translateY(-2px);
          box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
        }

        .btn-join:disabled {
          background: #9ca3af;
          cursor: not-allowed;
          opacity: 0.6;
        }

        .btn-leave {
          background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
          color: white;
          padding: 1rem 2rem;
          border: none;
          border-radius: 8px;
          cursor: pointer;
          font-size: 1.1rem;
          font-weight: 700;
          transition: all 0.3s;
          flex: 1;
          min-width: 200px;
        }

        .btn-leave:hover {
          transform: translateY(-2px);
          box-shadow: 0 6px 20px rgba(239, 68, 68, 0.4);
        }

        .btn-calendar {
          background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
          color: white;
          padding: 1rem 2rem;
          border: none;
          border-radius: 8px;
          cursor: pointer;
          font-size: 1.1rem;
          font-weight: 700;
          transition: all 0.3s;
          flex: 1;
          min-width: 200px;
        }

        .btn-calendar:hover {
          transform: translateY(-2px);
          box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4);
        }

        .btn-maps {
          display: inline-block;
          margin-top: 1rem;
          padding: 0.75rem 1.5rem;
          background: linear-gradient(135deg, #34a853 0%, #0f9d58 100%);
          color: white;
          border: none;
          border-radius: 8px;
          cursor: pointer;
          font-size: 1rem;
          font-weight: 600;
          text-decoration: none;
          transition: all 0.3s;
          width: 100%;
          text-align: center;
          box-shadow: 0 2px 8px rgba(52, 168, 83, 0.3);
        }

        .btn-maps:hover {
          transform: translateY(-2px);
          box-shadow: 0 6px 20px rgba(52, 168, 83, 0.4);
        }

        .map-container {
          border-radius: 8px;
          overflow: hidden;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }
      `}</style>
    </div>
  )
}
