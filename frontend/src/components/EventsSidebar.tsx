import { Event, CATEGORIES } from '../types'
import { useNavigate } from 'react-router-dom'
import { getLanguagesDisplay } from '../languages'
import { sanitizeText } from '../utils/sanitize'
import './EventsSidebar.css'

interface EventsSidebarProps {
  events: Event[]
  isOpen: boolean
  onClose: () => void
  onEventClick: (event: Event) => void
}

function EventsSidebar({ events, onEventClick }: EventsSidebarProps) {
  const navigate = useNavigate()

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
    <div className="events-sidebar">
      <div className="sidebar-header">
        <h2>{events.length} Events</h2>
      </div>

        <div className="sidebar-content">
          {events.length === 0 ? (
            <div className="no-events">
              <p>No events found in this searching area</p>
              <p className="hint">Try widening your search area by zooming out or moving the map!</p>
              <p className="hint">Or be the first to create an event here! 🎉</p>
            </div>
          ) : (
            events.map((event) => (
              <div
                key={event.id}
                className="event-card"
                onClick={() => {
                  if (event.slug) {
                    navigate(`/event/${event.slug}`)
                  } else {
                    onEventClick(event)
                  }
                }}
              >
                <div className="event-category">
                  {CATEGORIES[event.category as keyof typeof CATEGORIES]}
                </div>
                <h3 className="event-title" dangerouslySetInnerHTML={{ __html: sanitizeText(event.title) }} />
                <p className="event-description" dangerouslySetInnerHTML={{ __html: sanitizeText(event.description) }} />

                <div className="event-details">
                  <div className="event-meta">
                    <span className="meta-label">📅</span>
                    <span>{formatDateTime(event.start_time)}</span>
                  </div>

                  {event.max_participants && (
                    <div className="event-meta">
                      <span className="meta-label">👥</span>
                      <span>Max {event.max_participants} participants</span>
                    </div>
                  )}

                  {event.participant_count !== undefined && (
                    <div className="event-meta">
                      <span className="meta-label">✓</span>
                      <span>{event.participant_count} joined{event.max_participants ? `/${event.max_participants}` : ''}</span>
                    </div>
                  )}

                  <div className="event-meta">
                    <span className="meta-label">👤</span>
                    <span>
                      Created by{' '}
                      <span
                        className="creator-link"
                        onClick={(e) => {
                          e.stopPropagation()
                          if (event.user_id) navigate(`/profile/${event.user_id}`)
                        }}
                        style={{ cursor: 'pointer', color: '#667eea', textDecoration: 'underline', fontWeight: 600 }}
                      >
                        {event.creator_name}
                      </span>
                    </span>
                  </div>

                  {event.event_languages && (
                    <div className="event-meta">
                      <span className="meta-label">🗣️</span>
                      <span>{getLanguagesDisplay(event.event_languages)}</span>
                    </div>
                  )}

                  {event.creator_languages && (
                    <div className="event-meta">
                      <span className="meta-label">🌍</span>
                      <span>{event.creator_languages.split(',').map(code => {
                        const langNames: { [key: string]: string } = {
                          'de': 'German', 'fr': 'French', 'it': 'Italian', 'rm': 'Romansh',
                          'en': 'English', 'es': 'Spanish', 'pt': 'Portuguese', 'pl': 'Polish',
                          'tr': 'Turkish', 'ar': 'Arabic'
                        }
                        return langNames[code] || code
                      }).join(', ')}</span>
                    </div>
                  )}
                </div>

                <div className="event-restrictions">
                  {event.gender_restriction && event.gender_restriction !== 'any' && (
                    <span className="restriction-badge">{event.gender_restriction}</span>
                  )}
                  {event.age_min > 0 && event.age_max < 99 && (
                    <span className="restriction-badge">
                      {event.age_min}-{event.age_max} years
                    </span>
                  )}
                  {event.smoking_allowed && (
                    <span className="restriction-badge">🚬 Smoking OK</span>
                  )}
                  {event.alcohol_allowed && (
                    <span className="restriction-badge">🍺 Alcohol OK</span>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
    </div>
  )
}

export default EventsSidebar
