export interface User {
  id: number
  email: string
  name: string
  age?: number
  gender?: string
  bio?: string
  threema?: string
  languages?: string  // Comma-separated language codes (e.g., "en,de,fr")
  is_admin: boolean
  is_blocked: boolean
  email_verified: boolean
  created_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface Event {
  id?: number
  user_id?: number
  title: string
  description: string
  category: string
  latitude: number
  longitude: number
  start_time: string
  end_time?: string
  creator_name: string
  max_participants?: number
  gender_restriction: 'any' | 'male' | 'female' | 'non-binary'
  age_min: number
  age_max: number
  smoking_allowed: boolean
  alcohol_allowed: boolean
  event_languages?: string  // Comma-separated language codes for the event
  slug?: string
  created_at?: string
  user_email?: string
  creator_languages?: string  // Comma-separated language codes from creator's profile
  participant_count?: number  // Number of users who joined this event

  // Privacy controls
  hide_organizer_until_joined: boolean
  hide_participants_until_joined: boolean
  require_verified_to_join: boolean
  require_verified_to_view: boolean
  is_participant?: boolean  // Whether current user is a participant
}

export const CATEGORIES = {
  'social_drinks': 'Social & Drinks 🍻',
  'sports_fitness': 'Sports & Fitness 🏃',
  'food_dining': 'Food & Dining 🍕',
  'business_networking': 'Business & Networking 💼',
  'gaming_hobbies': 'Gaming & Hobbies 🎮',
  'learning_skills': 'Learning & Skills 📚',
  'adventure_travel': 'Adventure & Travel ✈️',
  'parents_kids': 'Parents & Kids 👶',
} as const

export type CategoryKey = keyof typeof CATEGORIES
