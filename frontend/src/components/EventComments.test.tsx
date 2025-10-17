import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import EventComments from './EventComments'

vi.mock('axios')

describe('EventComments Component', () => {
  const mockComments = [
    {
      id: 1,
      event_id: 1,
      user_id: 1,
      user_name: 'John Doe',
      comment: 'Great event!',
      created_at: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
      is_deleted: false,
      is_own: true
    },
    {
      id: 2,
      event_id: 1,
      user_id: 2,
      user_name: 'Jane Smith',
      comment: 'Looking forward to it!',
      created_at: new Date(Date.now() - 7200000).toISOString(), // 2 hours ago
      is_deleted: false,
      is_own: false
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.setItem('token', 'mock-token')
  })

  describe('Non-participant view', () => {
    it('should show join message when not a participant', () => {
      render(<EventComments eventId={1} isParticipant={false} />)

      expect(screen.getByText('💬 Comments')).toBeInTheDocument()
      expect(screen.getByText('Join this event to see and post comments')).toBeInTheDocument()
    })

    it('should not fetch comments when not a participant', () => {
      render(<EventComments eventId={1} isParticipant={false} />)

      expect(axios.get).not.toHaveBeenCalled()
    })
  })

  describe('Participant view', () => {
    it('should show loading state initially', () => {
      vi.mocked(axios.get).mockImplementation(() => new Promise(() => {}))

      render(<EventComments eventId={1} isParticipant={true} />)

      expect(screen.getByText('Loading comments...')).toBeInTheDocument()
    })

    it('should fetch and display comments', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockComments })

      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('💬 Comments (2)')).toBeInTheDocument()
      })

      expect(screen.getByText('Great event!')).toBeInTheDocument()
      expect(screen.getByText('Looking forward to it!')).toBeInTheDocument()
    })

    it('should show error when comments fail to load', async () => {
      vi.mocked(axios.get).mockRejectedValue({
        response: { status: 500 }
      })

      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('Failed to load comments')).toBeInTheDocument()
      })
    })

    it('should show specific error for 403 forbidden', async () => {
      vi.mocked(axios.get).mockRejectedValue({
        response: { status: 403 }
      })

      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('Only participants can view comments')).toBeInTheDocument()
      })
    })

    it('should show empty state when no comments', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [] })

      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('No comments yet. Be the first to comment!')).toBeInTheDocument()
      })
    })
  })

  describe('Comment posting', () => {
    it('should post new comment successfully', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [] })
      vi.mocked(axios.post).mockResolvedValue({ data: { success: true } })

      const user = userEvent.setup()
      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByPlaceholderText('Write a comment...')).toBeInTheDocument()
      })

      const textarea = screen.getByPlaceholderText('Write a comment...')
      await user.type(textarea, 'This is my comment')

      const postButton = screen.getByText('Post Comment')
      await user.click(postButton)

      await waitFor(() => {
        expect(axios.post).toHaveBeenCalledWith(
          '/api/events/1/comments',
          { comment: 'This is my comment' },
          { headers: { Authorization: 'Bearer mock-token' } }
        )
      })
    })

    it('should show character count', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [] })

      const user = userEvent.setup()
      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('0/1000')).toBeInTheDocument()
      })

      const textarea = screen.getByPlaceholderText('Write a comment...')
      await user.type(textarea, 'Test')

      expect(screen.getByText('4/1000')).toBeInTheDocument()
    })

    it('should disable submit when comment is empty', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [] })

      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        const postButton = screen.getByText('Post Comment')
        expect(postButton).toBeDisabled()
      })
    })

    it('should clear textarea after successful post', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: [] })
      vi.mocked(axios.post).mockResolvedValue({ data: { success: true } })

      const user = userEvent.setup()
      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByPlaceholderText('Write a comment...')).toBeInTheDocument()
      })

      const textarea = screen.getByPlaceholderText('Write a comment...') as HTMLTextAreaElement
      await user.type(textarea, 'Test comment')
      await user.click(screen.getByText('Post Comment'))

      await waitFor(() => {
        expect(textarea.value).toBe('')
      })
    })
  })

  describe('Comment editing', () => {
    it('should show edit UI when edit button is clicked', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockComments })

      const user = userEvent.setup()
      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('Great event!')).toBeInTheDocument()
      })

      const editButton = screen.getByText('✏️ Edit')
      await user.click(editButton)

      expect(screen.getByText('Cancel')).toBeInTheDocument()
      expect(screen.getByText('Save')).toBeInTheDocument()
    })

    it('should cancel edit when cancel button is clicked', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockComments })

      const user = userEvent.setup()
      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('Great event!')).toBeInTheDocument()
      })

      await user.click(screen.getByText('✏️ Edit'))
      await user.click(screen.getByText('Cancel'))

      expect(screen.queryByText('Cancel')).not.toBeInTheDocument()
    })

    it('should save edited comment successfully', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockComments })
      vi.mocked(axios.put).mockResolvedValue({ data: { success: true } })

      const user = userEvent.setup()
      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('Great event!')).toBeInTheDocument()
      })

      await user.click(screen.getByText('✏️ Edit'))

      const textarea = screen.getAllByRole('textbox')[1] as HTMLTextAreaElement
      await user.clear(textarea)
      await user.type(textarea, 'Updated comment')

      await user.click(screen.getByText('Save'))

      await waitFor(() => {
        expect(axios.put).toHaveBeenCalledWith(
          '/api/comments/1',
          { comment: 'Updated comment' },
          { headers: { Authorization: 'Bearer mock-token' } }
        )
      })
    })
  })

  describe('Comment deletion', () => {
    it('should delete comment after confirmation', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockComments })
      vi.mocked(axios.delete).mockResolvedValue({ data: { success: true } })

      window.confirm = vi.fn(() => true)

      const user = userEvent.setup()
      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('Great event!')).toBeInTheDocument()
      })

      await user.click(screen.getByText('🗑️ Delete'))

      expect(window.confirm).toHaveBeenCalledWith('Delete this comment?')
      await waitFor(() => {
        expect(axios.delete).toHaveBeenCalledWith(
          '/api/comments/1',
          { headers: { Authorization: 'Bearer mock-token' } }
        )
      })
    })

    it('should not delete comment when confirmation is cancelled', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockComments })

      window.confirm = vi.fn(() => false)

      const user = userEvent.setup()
      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('Great event!')).toBeInTheDocument()
      })

      await user.click(screen.getByText('🗑️ Delete'))

      expect(axios.delete).not.toHaveBeenCalled()
    })
  })

  describe('Time formatting', () => {
    it('should show "just now" for recent comments', async () => {
      const recentComment = [{
        ...mockComments[0],
        created_at: new Date().toISOString()
      }]

      vi.mocked(axios.get).mockResolvedValue({ data: recentComment })

      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('just now')).toBeInTheDocument()
      })
    })

    it('should show edited indicator', async () => {
      const editedComment = [{
        ...mockComments[0],
        updated_at: new Date().toISOString()
      }]

      vi.mocked(axios.get).mockResolvedValue({ data: editedComment })

      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText(/\(edited\)/)).toBeInTheDocument()
      })
    })
  })

  describe('Own comment identification', () => {
    it('should show "You" badge on own comments', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockComments })

      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        expect(screen.getByText('You')).toBeInTheDocument()
      })
    })

    it('should show edit/delete buttons only on own comments', async () => {
      vi.mocked(axios.get).mockResolvedValue({ data: mockComments })

      render(<EventComments eventId={1} isParticipant={true} />)

      await waitFor(() => {
        const editButtons = screen.getAllByText('✏️ Edit')
        const deleteButtons = screen.getAllByText('🗑️ Delete')

        // Only one comment is own, so only one set of buttons
        expect(editButtons).toHaveLength(1)
        expect(deleteButtons).toHaveLength(1)
      })
    })
  })
})
