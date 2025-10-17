import axios from 'axios'
import { Event, User } from './types'
import { API_BASE_URL } from './config'

// Log API URL in development for debugging
if (import.meta.env.DEV) {
  console.log('API Base URL:', API_BASE_URL)
}

interface Place {
  display_name: string
  lat: string
  lon: string
  type: string
  importance: number
}

export const api = {
  getEvents: async (params?: any): Promise<Event[]> => {
    const response = await axios.get(`${API_BASE_URL}/events`, { params })
    return response.data
  },

  getEvent: async (id: number): Promise<Event> => {
    const response = await axios.get(`${API_BASE_URL}/events/${id}`)
    return response.data
  },

  createEvent: async (event: Event): Promise<Event> => {
    const response = await axios.post(`${API_BASE_URL}/events`, event)
    return response.data
  },

  updateEvent: async (id: number, event: Event): Promise<Event> => {
    const response = await axios.put(`${API_BASE_URL}/events/${id}`, event)
    return response.data
  },

  deleteEvent: async (id: number): Promise<void> => {
    await axios.delete(`${API_BASE_URL}/events/${id}`)
  },

  // Event participation
  joinEvent: async (eventId: number): Promise<void> => {
    await axios.post(`${API_BASE_URL}/events/${eventId}/join`)
  },

  leaveEvent: async (eventId: number): Promise<void> => {
    await axios.delete(`${API_BASE_URL}/events/${eventId}/leave`)
  },

  getEventParticipants: async (eventId: number): Promise<User[]> => {
    const response = await axios.get(`${API_BASE_URL}/events/${eventId}/participants`)
    return response.data
  },

  getPublicEvent: async (slug: string): Promise<Event> => {
    const response = await axios.get(`${API_BASE_URL}/public/events/${slug}`)
    return response.data
  },

  // Place search
  searchPlaces: async (query: string): Promise<Place[]> => {
    const response = await axios.get(`${API_BASE_URL}/search/places`, {
      params: { q: query }
    })
    return response.data
  },

  // Email verification
  verifyEmail: async (token: string): Promise<void> => {
    await axios.get(`${API_BASE_URL}/auth/verify-email`, {
      params: { token }
    })
  },

  resendVerificationEmail: async (email: string): Promise<void> => {
    await axios.post(`${API_BASE_URL}/auth/resend-verification`, { email })
  },

  // Password reset
  forgotPassword: async (email: string): Promise<void> => {
    await axios.post(`${API_BASE_URL}/auth/forgot-password`, { email })
  },

  resetPassword: async (token: string, newPassword: string): Promise<void> => {
    await axios.post(`${API_BASE_URL}/auth/reset-password`, {
      token,
      new_password: newPassword
    })
  },
}
