import { useState, FormEvent, useEffect } from 'react'
import { MapContainer, TileLayer, Marker } from 'react-leaflet'
import { api } from '../api'
import { Event, CATEGORIES } from '../types'
import { useAuth } from '../AuthContext'
import { LANGUAGES, getLanguagesDisplay } from '../languages'
import 'leaflet/dist/leaflet.css'
import './EventForm.css'

interface EventFormProps {
  initialLocation: { lat: number; lng: number } | null
  onClose: () => void
  event?: Event // For editing existing events
}

interface Place {
  display_name: string
  lat: string
  lon: string
}

function EventForm({ initialLocation, onClose, event }: EventFormProps) {
  const { user } = useAuth()

  const isEditMode = !!event

  const [formData, setFormData] = useState({
    title: event?.title || '',
    description: event?.description || '',
    category: event?.category || 'social_drinks',
    latitude: event?.latitude || initialLocation?.lat || 0,
    longitude: event?.longitude || initialLocation?.lng || 0,
    start_time: event?.start_time ? new Date(event.start_time).toISOString().slice(0, 16) : '',
    end_time: event?.end_time ? new Date(event.end_time).toISOString().slice(0, 16) : '',
    creator_name: event?.creator_name || user?.name || '',
    max_participants: event?.max_participants?.toString() || '',
    gender_restriction: event?.gender_restriction || 'any',
    age_min: event?.age_min || 0,
    age_max: event?.age_max || 99,
    smoking_allowed: event?.smoking_allowed || false,
    alcohol_allowed: event?.alcohol_allowed || false,
    // Auto-populate from user's profile languages when creating new event
    event_languages: event?.event_languages || user?.languages || '',
    // Privacy settings (defaults - all enabled for maximum privacy)
    hide_organizer_until_joined: event?.hide_organizer_until_joined !== undefined ? event.hide_organizer_until_joined : true,
    hide_participants_until_joined: event?.hide_participants_until_joined !== undefined ? event.hide_participants_until_joined : true,
    require_verified_to_join: event?.require_verified_to_join !== undefined ? event.require_verified_to_join : false,
    require_verified_to_view: event?.require_verified_to_view !== undefined ? event.require_verified_to_view : true,
    allow_unregistered_users: event?.allow_unregistered_users !== undefined ? event.allow_unregistered_users : false,
  })

  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [placeSearch, setPlaceSearch] = useState('')
  const [places, setPlaces] = useState<Place[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const [hasSelectedPlace, setHasSelectedPlace] = useState(false)
  const [validationErrors, setValidationErrors] = useState<{ [key: string]: string }>({})

  // Validation function
  const validateField = (name: string, value: any): string => {
    switch (name) {
      case 'title':
        if (typeof value === 'string') {
          if (value.length < 3) return 'Title must be at least 3 characters'
          if (value.length > 200) return 'Title must be less than 200 characters'
        }
        break
      case 'description':
        if (typeof value === 'string') {
          if (value.length < 10) return 'Description must be at least 10 characters'
          if (value.length > 5000) return 'Description must be less than 5000 characters'
        }
        break
      case 'age_min':
      case 'age_max':
        if (typeof value === 'number') {
          if (value < 0 || value > 150) return 'Age must be between 0 and 150'
        }
        break
    }
    return ''
  }

  // Validate form on submit
  const validateForm = (): boolean => {
    const errors: { [key: string]: string } = {}

    // Title validation
    const titleError = validateField('title', formData.title)
    if (titleError) errors.title = titleError

    // Description validation
    const descError = validateField('description', formData.description)
    if (descError) errors.description = descError

    // Age range validation
    if (formData.age_min > formData.age_max) {
      errors.age_min = 'Minimum age cannot be greater than maximum age'
    }

    // Max participants validation
    if (formData.max_participants && parseInt(formData.max_participants) < 0) {
      errors.max_participants = 'Max participants must be 0 or positive'
    }

    setValidationErrors(errors)
    return Object.keys(errors).length === 0
  }

  // Search for places when user types
  useEffect(() => {
    // Don't show autocomplete if user has already selected a place
    if (hasSelectedPlace) {
      setPlaces([])
      return
    }

    // Require minimum 3 characters to start search (fair usage)
    if (placeSearch.length < 3) {
      setPlaces([])
      return
    }

    // Debounce for 800ms to reduce API calls (fair usage of Photon API)
    const timer = setTimeout(async () => {
      setIsSearching(true)
      try {
        const results = await api.searchPlaces(placeSearch)
        setPlaces(results || [])
      } catch (err) {
        console.error('Failed to search places:', err)
        setPlaces([])
      } finally {
        setIsSearching(false)
      }
    }, 800)

    return () => clearTimeout(timer)
  }, [placeSearch, hasSelectedPlace])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')

    // Validate form before submission
    if (!validateForm()) {
      setError('Please fix the validation errors before submitting')
      return
    }

    setIsSubmitting(true)

    try {
      const eventData: Event = {
        ...formData,
        max_participants: formData.max_participants ? parseInt(formData.max_participants) : undefined,
      }

      if (isEditMode && event?.id) {
        await api.updateEvent(event.id, eventData)
      } else {
        await api.createEvent(eventData)
      }
      onClose()
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to ${isEditMode ? 'update' : 'create'} event`)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target

    // Clear validation error for this field when user starts typing
    if (validationErrors[name]) {
      setValidationErrors(prev => {
        const newErrors = { ...prev }
        delete newErrors[name]
        return newErrors
      })
    }

    if (type === 'checkbox') {
      const checked = (e.target as HTMLInputElement).checked
      setFormData(prev => ({ ...prev, [name]: checked }))
    } else if (type === 'number') {
      const numValue = parseFloat(value) || 0
      setFormData(prev => ({ ...prev, [name]: numValue }))

      // Validate number fields on change for immediate feedback
      const error = validateField(name, numValue)
      if (error) {
        setValidationErrors(prev => ({ ...prev, [name]: error }))
      }
    } else {
      setFormData(prev => ({ ...prev, [name]: value }))
    }
  }

  const handlePlaceSelect = (place: Place) => {
    setFormData(prev => ({
      ...prev,
      latitude: parseFloat(place.lat),
      longitude: parseFloat(place.lon),
    }))
    setPlaceSearch(place.display_name)
    setPlaces([])
    setHasSelectedPlace(true)
  }

  // Language selection helpers
  const toggleLanguage = (languageCode: string) => {
    const currentLanguages = formData.event_languages ? formData.event_languages.split(',') : []
    const updatedLanguages = currentLanguages.includes(languageCode)
      ? currentLanguages.filter(code => code !== languageCode)
      : [...currentLanguages, languageCode]

    setFormData(prev => ({
      ...prev,
      event_languages: updatedLanguages.join(',')
    }))
  }

  const isLanguageSelected = (code: string) => {
    return formData.event_languages ? formData.event_languages.split(',').includes(code) : false
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>{isEditMode ? 'Edit Event' : 'Create New Event'}</h2>
          <button className="close-button" onClick={onClose}>&times;</button>
        </div>

        <form onSubmit={handleSubmit} className="event-form">
          {error && <div className="error-message">{error}</div>}

          <div className="form-section">
            <h3>Event Details</h3>

            <div className="form-group">
              <label htmlFor="title">Event Title * (3-200 characters)</label>
              <input
                type="text"
                id="title"
                name="title"
                value={formData.title}
                onChange={handleChange}
                required
                minLength={3}
                maxLength={200}
                placeholder="e.g., Coffee & Chat"
                className={validationErrors.title ? 'error' : ''}
              />
              {validationErrors.title && (
                <span className="field-error">{validationErrors.title}</span>
              )}
              <small className="character-count">
                {formData.title.length}/200 characters
              </small>
            </div>

            <div className="form-group">
              <label htmlFor="description">Description * (10-5000 characters)</label>
              <textarea
                id="description"
                name="description"
                value={formData.description}
                onChange={handleChange}
                required
                minLength={10}
                maxLength={5000}
                rows={4}
                placeholder="Describe your event, what you'll do, and what to expect... (minimum 10 characters)"
                className={validationErrors.description ? 'error' : ''}
              />
              {validationErrors.description && (
                <span className="field-error">{validationErrors.description}</span>
              )}
              <small className="character-count">
                {formData.description.length}/5000 characters
                {formData.description.length < 10 && (
                  <span className="warning"> - {10 - formData.description.length} more needed</span>
                )}
              </small>
            </div>

            <div className="form-group">
              <label htmlFor="category">Category *</label>
              <select
                id="category"
                name="category"
                value={formData.category}
                onChange={handleChange}
                required
              >
                {Object.entries(CATEGORIES).map(([key, label]) => (
                  <option key={key} value={key}>
                    {label}
                  </option>
                ))}
              </select>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label htmlFor="start_time">Start Time *</label>
                <input
                  type="datetime-local"
                  id="start_time"
                  name="start_time"
                  value={formData.start_time}
                  onChange={handleChange}
                  required
                />
              </div>

              <div className="form-group">
                <label htmlFor="end_time">End Time (optional)</label>
                <input
                  type="datetime-local"
                  id="end_time"
                  name="end_time"
                  value={formData.end_time}
                  onChange={handleChange}
                />
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="max_participants">Max Participants (optional)</label>
              <input
                type="number"
                id="max_participants"
                name="max_participants"
                value={formData.max_participants}
                onChange={handleChange}
                min="1"
                placeholder="Leave empty for unlimited"
              />
            </div>
          </div>

          <div className="form-section">
            <h3>Location</h3>

            <div className="place-search-container">
              <div className="form-group">
                <label htmlFor="place_search">Search for a Place</label>
                <input
                  type="text"
                  id="place_search"
                  value={placeSearch}
                  onChange={(e) => {
                    setPlaceSearch(e.target.value)
                    // Allow user to search again if they modify the selected place
                    if (hasSelectedPlace) {
                      setHasSelectedPlace(false)
                    }
                  }}
                  placeholder="e.g., Starbucks Warsaw, Central Park NYC..."
                />
              </div>

              {isSearching && (
                <p className="form-hint">Searching...</p>
              )}

              {places.length > 0 && (
                <div className="place-search-results">
                  {places.map((place, index) => (
                    <div
                      key={index}
                      className="place-result"
                      onClick={() => handlePlaceSelect(place)}
                    >
                      {place.display_name}
                    </div>
                  ))}
                </div>
              )}
            </div>

            {formData.latitude !== 0 && formData.longitude !== 0 && (
              <div className="location-preview">
                <label>Event Location Preview</label>
                <div className="location-map-preview">
                  <MapContainer
                    center={[formData.latitude, formData.longitude]}
                    zoom={15}
                    style={{ height: '250px', width: '100%', borderRadius: '12px' }}
                    scrollWheelZoom={false}
                    dragging={false}
                    zoomControl={false}
                    doubleClickZoom={false}
                    touchZoom={false}
                  >
                    <TileLayer
                      attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
                      url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                    />
                    <Marker position={[formData.latitude, formData.longitude]} />
                  </MapContainer>
                </div>
                <p className="form-hint">📍 Your event will be here. You can search above or click the main map to change location.</p>
              </div>
            )}

            {(formData.latitude === 0 || formData.longitude === 0) && (
              <p className="form-hint">💡 Search for a place above or click on the main map to set event location</p>
            )}
          </div>

          <div className="form-section">
            <h3>Contact Information</h3>

            <div className="form-group">
              <label htmlFor="creator_name">Your Name *</label>
              <input
                type="text"
                id="creator_name"
                name="creator_name"
                value={formData.creator_name}
                onChange={handleChange}
                required
                placeholder="How should people call you?"
              />
            </div>
          </div>

          <div className="form-section">
            <h3>Preferences & Restrictions</h3>

            <div className="form-group">
              <label htmlFor="gender_restriction">Gender Preference</label>
              <select
                id="gender_restriction"
                name="gender_restriction"
                value={formData.gender_restriction}
                onChange={handleChange}
              >
                <option value="any">Anyone welcome</option>
                <option value="male">Male only</option>
                <option value="female">Female only</option>
                <option value="non-binary">Non-binary only</option>
              </select>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label htmlFor="age_min">Minimum Age</label>
                <input
                  type="number"
                  id="age_min"
                  name="age_min"
                  value={formData.age_min}
                  onChange={handleChange}
                  min="0"
                  max="99"
                />
              </div>

              <div className="form-group">
                <label htmlFor="age_max">Maximum Age</label>
                <input
                  type="number"
                  id="age_max"
                  name="age_max"
                  value={formData.age_max}
                  onChange={handleChange}
                  min="0"
                  max="99"
                />
              </div>
            </div>

            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  name="smoking_allowed"
                  checked={formData.smoking_allowed}
                  onChange={handleChange}
                />
                <span>🚬 Smoking is allowed</span>
              </label>
            </div>

            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  name="alcohol_allowed"
                  checked={formData.alcohol_allowed}
                  onChange={handleChange}
                />
                <span>🍺 Alcohol is allowed</span>
              </label>
            </div>

            <p className="form-hint">
              Set your comfort level. Being specific helps people find events that match their preferences.
            </p>
          </div>

          <div className="form-section">
            <h3>Event Languages</h3>

            <div className="form-group">
              <label htmlFor="event_languages">
                Which languages will be spoken at this event?
                <span className="optional-label"> (Select all that apply)</span>
              </label>
              <small className="form-hint">
                Help participants find events where they can communicate comfortably
              </small>
              <div className="language-selector">
                {LANGUAGES.map(lang => (
                  <button
                    key={lang.code}
                    type="button"
                    className={`language-option ${isLanguageSelected(lang.code) ? 'selected' : ''}`}
                    onClick={() => toggleLanguage(lang.code)}
                  >
                    <span className="language-flag">{lang.flag}</span>
                    <span className="language-name">{lang.name}</span>
                  </button>
                ))}
              </div>
              {formData.event_languages && (
                <div className="selected-languages">
                  <strong>Selected:</strong> {getLanguagesDisplay(formData.event_languages)}
                </div>
              )}
            </div>
          </div>

          <div className="form-section privacy-section">
            <h3>🔒 Privacy & Access Control</h3>
            <p className="form-hint">
              Control who can see event information and who can join
            </p>

            <div className="privacy-info-box">
              ℹ️ <strong>Automatic Protection:</strong> All events require verified email to join (platform-wide security). Your contact information is automatically hidden from unverified users. Only verified users who join your event will see your full contact details.
            </div>

            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  name="hide_organizer_until_joined"
                  checked={formData.hide_organizer_until_joined}
                  onChange={handleChange}
                />
                <span>
                  <strong>Hide my name and contact until someone joins</strong>
                  <small className="checkbox-description">
                    Your name becomes "Join to see organizer" and contact stays hidden for all users (even verified) until they join
                  </small>
                </span>
              </label>
            </div>

            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  name="hide_participants_until_joined"
                  checked={formData.hide_participants_until_joined}
                  onChange={handleChange}
                />
                <span>
                  <strong>Hide participant list until someone joins</strong>
                  <small className="checkbox-description">
                    Only show participant count publicly, full list visible to participants
                  </small>
                </span>
              </label>
            </div>

            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  name="require_verified_to_view"
                  checked={formData.require_verified_to_view}
                  onChange={handleChange}
                />
                <span>
                  <strong>Require verified email to view event details</strong>
                  <small className="checkbox-description">
                    Hide event completely from unverified users (they won't see it on the map)
                  </small>
                </span>
              </label>
            </div>

            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  name="allow_unregistered_users"
                  checked={formData.allow_unregistered_users}
                  onChange={handleChange}
                />
                <span>
                  <strong>Allow unregistered users to view this event</strong>
                  <small className="checkbox-description">
                    Anyone can see this event without creating an account (recommended for public events)
                  </small>
                </span>
              </label>
            </div>

            <div className="privacy-recommendation">
              💡 <strong>Recommended:</strong> Enable "Hide organizer" for extra privacy protection
            </div>
          </div>

          <div className="form-actions">
            <button type="button" className="button-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="button-primary" disabled={isSubmitting}>
              {isSubmitting ? (isEditMode ? 'Updating...' : 'Creating...') : (isEditMode ? 'Update Event' : 'Create Event')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default EventForm
