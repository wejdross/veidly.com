import { Event, User } from '../types'

export const mockUser: User = {
  id: 1,
  email: 'test@example.com',
  name: 'Test User',
  bio: 'Test bio',
    threema: 'TESTID',
  languages: 'en,de',
  is_admin: false,
  is_blocked: false,
  created_at: '2024-01-01T00:00:00Z',
}

export const mockAdmin: User = {
  ...mockUser,
  id: 2,
  email: 'admin@example.com',
  name: 'Admin User',
  is_admin: true,
}

export const mockEvent: Event = {
  id: 1,
  user_id: 1,
  title: 'Test Event',
  description: 'This is a test event',
  category: 'social_drinks',
  latitude: 52.52,
  longitude: 13.405,
  start_time: '2024-12-25T18:00:00Z',
  end_time: '2024-12-25T22:00:00Z',
  creator_name: 'Test Creator',
  max_participants: 10,
  gender_restriction: 'any',
  age_min: 18,
  age_max: 99,
  smoking_allowed: false,
  alcohol_allowed: true,
  slug: 'test-event-abc123',
  created_at: '2024-01-01T00:00:00Z',
  participant_count: 5,
  creator_languages: 'en,de',
}

export const mockEvents: Event[] = [
  mockEvent,
  {
    ...mockEvent,
    id: 2,
    title: 'Another Event',
    description: 'Another test event',
    category: 'sports_fitness',
    latitude: 52.51,
    longitude: 13.39,
    slug: 'another-event-xyz789',
    participant_count: 3,
  },
  {
    ...mockEvent,
    id: 3,
    title: 'Third Event',
    description: 'Yet another event',
    category: 'cultural_arts',
    latitude: 52.53,
    longitude: 13.42,
    slug: 'third-event-def456',
    participant_count: 8,
  },
]

export const mockParticipant = {
  user_id: 1,
  name: 'Participant One',
  bio: 'Test bio',
  languages: 'en,de',
  joined_at: '2024-01-01T10:00:00Z',
}

export const mockParticipants = [
  mockParticipant,
  {
    ...mockParticipant,
    user_id: 2,
    name: 'Participant Two',
    languages: 'en',
    joined_at: '2024-01-01T11:00:00Z',
  },
]
