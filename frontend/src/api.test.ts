import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'
import { api } from './api'
import { mockEvent, mockUser, mockEvents } from './test/mockData'
import { API_BASE_URL } from './config'

// Mock axios with interceptors support
vi.mock('axios', () => {
  const mockAxios = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    defaults: {},
    interceptors: {
      request: {
        use: vi.fn(() => 0),
        eject: vi.fn(),
      },
      response: {
        use: vi.fn(() => 0),
        eject: vi.fn(),
      },
    },
  }
  return { default: mockAxios }
})

describe('API Module', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset mock implementations
    vi.mocked(axios.get).mockReset()
    vi.mocked(axios.post).mockReset()
    vi.mocked(axios.put).mockReset()
    vi.mocked(axios.delete).mockReset()
  })

  describe('getEvents', () => {
    it('should fetch events without params', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockEvents })

      const result = await api.getEvents()

      expect(axios.get).toHaveBeenCalledWith(`${API_BASE_URL}/events`, { params: undefined })
      expect(result).toEqual(mockEvents)
    })

    it('should fetch events with filter params', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockEvents })

      const params = { category: 'social_drinks', gender: 'any' }
      const result = await api.getEvents(params)

      expect(axios.get).toHaveBeenCalledWith(`${API_BASE_URL}/events`, { params })
      expect(result).toEqual(mockEvents)
    })
  })

  describe('getEvent', () => {
    it('should fetch a single event by id', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockEvent })

      const result = await api.getEvent(1)

      expect(axios.get).toHaveBeenCalledWith(`${API_BASE_URL}/events/1`)
      expect(result).toEqual(mockEvent)
    })
  })

  describe('createEvent', () => {
    it('should create a new event', async () => {
      vi.mocked(axios.post).mockResolvedValue({ data: mockEvent })

      const result = await api.createEvent(mockEvent)

      expect(axios.post).toHaveBeenCalledWith(`${API_BASE_URL}/events`, mockEvent)
      expect(result).toEqual(mockEvent)
    })
  })

  describe('updateEvent', () => {
    it('should update an existing event', async () => {
      const updatedEvent = { ...mockEvent, title: 'Updated Event' }
      vi.mocked(axios.put).mockResolvedValue({ data: updatedEvent })

      const result = await api.updateEvent(1, updatedEvent)

      expect(axios.put).toHaveBeenCalledWith(`${API_BASE_URL}/events/1`, updatedEvent)
      expect(result).toEqual(updatedEvent)
    })
  })

  describe('deleteEvent', () => {
    it('should delete an event', async () => {
      vi.mocked(axios.delete).mockResolvedValue({})

      await api.deleteEvent(1)

      expect(axios.delete).toHaveBeenCalledWith(`${API_BASE_URL}/events/1`)
    })
  })

  describe('joinEvent', () => {
    it('should join an event', async () => {
      vi.mocked(axios.post).mockResolvedValue({})

      await api.joinEvent(1)

      expect(axios.post).toHaveBeenCalledWith(`${API_BASE_URL}/events/1/join`)
    })
  })

  describe('leaveEvent', () => {
    it('should leave an event', async () => {
      vi.mocked(axios.delete).mockResolvedValue({})

      await api.leaveEvent(1)

      expect(axios.delete).toHaveBeenCalledWith(`${API_BASE_URL}/events/1/leave`)
    })
  })

  describe('getEventParticipants', () => {
    it('should fetch event participants', async () => {
      const participants = [mockUser]
      vi.mocked(axios.get).mockResolvedValue({ data: participants })

      const result = await api.getEventParticipants(1)

      expect(axios.get).toHaveBeenCalledWith(`${API_BASE_URL}/events/1/participants`)
      expect(result).toEqual(participants)
    })
  })
})
