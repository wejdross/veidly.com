import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { User } from '../types'
import { api } from '../api'
import './ParticipantModal.css'

interface ParticipantModalProps {
  eventId: number
  eventTitle: string
  onClose: () => void
}

function ParticipantModal({ eventId, eventTitle, onClose }: ParticipantModalProps) {
  const navigate = useNavigate()
  const [participants, setParticipants] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const availableLanguages: { [key: string]: string } = {
    'de': 'German',
    'fr': 'French',
    'it': 'Italian',
    'rm': 'Romansh',
    'en': 'English',
    'es': 'Spanish',
    'pt': 'Portuguese',
    'pl': 'Polish',
    'tr': 'Turkish',
    'ar': 'Arabic'
  }

  useEffect(() => {
    loadParticipants()
  }, [eventId])

  const loadParticipants = async () => {
    try {
      setLoading(true)
      const data = await api.getEventParticipants(eventId)
      setParticipants(data)
      setLoading(false)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load participants')
      setLoading(false)
    }
  }

  const handleProfileClick = (userId: number) => {
    onClose()
    navigate(`/profile/${userId}`)
  }

  const getLanguageNames = (languageCodes: string) => {
    return languageCodes
      .split(',')
      .map(code => availableLanguages[code] || code)
      .join(', ')
  }

  return (
    <>
      <div className="modal-overlay" onClick={onClose} />
      <div className="participant-modal">
        <div className="modal-header">
          <h2>Event Participants</h2>
          <button className="close-button" onClick={onClose}>✕</button>
        </div>

        <div className="modal-event-title">
          <strong>{eventTitle}</strong>
        </div>

        <div className="modal-content">
          {loading ? (
            <div className="loading-state">Loading participants...</div>
          ) : error ? (
            <div className="error-state">{error}</div>
          ) : participants.length === 0 ? (
            <div className="empty-state">
              <p>No participants yet</p>
              <p className="hint">Be the first to join this event!</p>
            </div>
          ) : (
            <div className="participants-list">
              {participants.map((participant) => (
                <div
                  key={participant.id}
                  className="participant-card"
                  onClick={() => handleProfileClick(participant.id)}
                >
                  <div className="participant-avatar">
                    {participant.name.charAt(0).toUpperCase()}
                  </div>
                  <div className="participant-info">
                    <h3>{participant.name}</h3>
                    {participant.bio && (
                      <p className="participant-bio">{participant.bio}</p>
                    )}
                    {participant.languages && (
                      <div className="participant-languages">
                        <span className="language-icon">🌍</span>
                        <span>{getLanguageNames(participant.languages)}</span>
                      </div>
                    )}
                  </div>
                  <div className="view-profile-arrow">→</div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="modal-footer">
          <p className="participant-count">
            {participants.length} {participants.length === 1 ? 'person' : 'people'} attending
          </p>
        </div>
      </div>
    </>
  )
}

export default ParticipantModal
