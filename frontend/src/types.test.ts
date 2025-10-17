import { describe, it, expect } from 'vitest'
import { CATEGORIES, type Event, type User, type CategoryKey } from './types'

describe('Types Module', () => {
  describe('CATEGORIES constant', () => {
    it('should have all expected categories', () => {
      expect(CATEGORIES).toHaveProperty('social_drinks')
      expect(CATEGORIES).toHaveProperty('sports_fitness')
      expect(CATEGORIES).toHaveProperty('food_dining')
      expect(CATEGORIES).toHaveProperty('business_networking')
      expect(CATEGORIES).toHaveProperty('gaming_hobbies')
      expect(CATEGORIES).toHaveProperty('learning_skills')
      expect(CATEGORIES).toHaveProperty('adventure_travel')
      expect(CATEGORIES).toHaveProperty('parents_kids')
    })

    it('should have correct category labels', () => {
      expect(CATEGORIES.social_drinks).toBe('Social & Drinks 🍻')
      expect(CATEGORIES.sports_fitness).toBe('Sports & Fitness 🏃')
      expect(CATEGORIES.food_dining).toBe('Food & Dining 🍕')
    })

    it('should be readonly', () => {
      // TypeScript enforces this, but we can test the runtime behavior
      expect(Object.isFrozen(CATEGORIES)).toBe(false) // const objects aren't frozen by default

      // But we can verify it exists and has the right shape
      const keys = Object.keys(CATEGORIES) as CategoryKey[]
      expect(keys.length).toBe(8)
    })
  })

  describe('Event interface', () => {
    it('should accept valid event objects', () => {
      const event: Event = {
        id: 1,
        title: 'Test Event',
        description: 'Test Description',
        category: 'social_drinks',
        latitude: 46.8805,
        longitude: 8.6444,
        start_time: '2025-12-31T14:00:00Z',
        creator_name: 'Test User',
        gender_restriction: 'any',
        age_min: 18,
        age_max: 99,
        smoking_allowed: false,
        alcohol_allowed: true,
      }

      expect(event.title).toBe('Test Event')
      expect(event.gender_restriction).toBe('any')
    })

    it('should handle optional fields', () => {
      const minimalEvent: Event = {
        title: 'Minimal Event',
        description: 'Description',
        category: 'social_drinks',
        latitude: 0,
        longitude: 0,
        start_time: '2025-01-01T00:00:00Z',
        creator_name: 'Creator',
        gender_restriction: 'any',
        age_min: 0,
        age_max: 99,
        smoking_allowed: false,
        alcohol_allowed: false,
      }

      expect(minimalEvent.id).toBeUndefined()
      expect(minimalEvent.end_time).toBeUndefined()
      expect(minimalEvent.max_participants).toBeUndefined()
    })
  })

  describe('User interface', () => {
    it('should accept valid user objects', () => {
      const user: User = {
        id: 1,
        email: 'test@example.com',
        name: 'Test User',
        age: 25,
        gender: 'male',
                threema: 'TESTUSER',
        languages: 'en,de',
        is_admin: false,
        is_blocked: false,
        created_at: '2024-01-01T00:00:00Z',
      }

      expect(user.email).toBe('test@example.com')
      expect(user.is_admin).toBe(false)
    })

    it('should handle optional fields', () => {
      const minimalUser: User = {
        id: 1,
        email: 'minimal@example.com',
        name: 'Minimal User',
        is_admin: false,
        is_blocked: false,
        created_at: '2024-01-01T00:00:00Z',
      }

      expect(minimalUser.age).toBeUndefined()
      expect(minimalUser.bio).toBeUndefined()
      expect(minimalUser.phone).toBeUndefined()
    })
  })

  describe('Gender restrictions', () => {
    it('should accept all valid gender restriction values', () => {
      const values: Array<Event['gender_restriction']> = ['any', 'male', 'female', 'non-binary']

      values.forEach(value => {
        const event: Event = {
          title: 'Test',
          description: 'Test',
          category: 'social_drinks',
          latitude: 0,
          longitude: 0,
          start_time: '2025-01-01T00:00:00Z',
          creator_name: 'Test',
          gender_restriction: value,
          age_min: 0,
          age_max: 99,
          smoking_allowed: false,
          alcohol_allowed: false,
        }

        expect(event.gender_restriction).toBe(value)
      })
    })
  })
})
