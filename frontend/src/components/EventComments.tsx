import React, { useState, useEffect } from 'react'
import axios from 'axios'
import './EventComments.css'

interface Comment {
  id: number
  event_id: number
  user_id: number
  comment: string
  user_name: string
  created_at: string
  updated_at?: string
  is_deleted: boolean
  is_own: boolean
}

interface EventCommentsProps {
  eventId: number
  isParticipant: boolean
}

const EventComments: React.FC<EventCommentsProps> = ({ eventId, isParticipant }) => {
  const [comments, setComments] = useState<Comment[]>([])
  const [newComment, setNewComment] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editText, setEditText] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Fetch comments
  const fetchComments = async () => {
    try {
      const token = localStorage.getItem('token')
      const response = await axios.get(`/api/events/${eventId}/comments`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      setComments(response.data)
      setError('')
    } catch (err: any) {
      if (err.response?.status === 403) {
        setError('Only participants can view comments')
      } else {
        setError('Failed to load comments')
      }
    } finally {
      setLoading(false)
    }
  }

  // Initial load
  useEffect(() => {
    if (isParticipant) {
      fetchComments()
    }
  }, [eventId, isParticipant])

  // Polling every 10 seconds
  useEffect(() => {
    if (!isParticipant) return

    const interval = setInterval(() => {
      fetchComments()
    }, 10000)

    return () => clearInterval(interval)
  }, [eventId, isParticipant])

  // Post new comment
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newComment.trim() || submitting) return

    setSubmitting(true)
    try {
      const token = localStorage.getItem('token')
      await axios.post(
        `/api/events/${eventId}/comments`,
        { comment: newComment },
        { headers: { Authorization: `Bearer ${token}` } }
      )
      setNewComment('')
      await fetchComments()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to post comment')
    } finally {
      setSubmitting(false)
    }
  }

  // Start editing
  const startEdit = (comment: Comment) => {
    setEditingId(comment.id)
    setEditText(comment.comment)
  }

  // Cancel editing
  const cancelEdit = () => {
    setEditingId(null)
    setEditText('')
  }

  // Save edit
  const saveEdit = async (commentId: number) => {
    if (!editText.trim() || submitting) return

    setSubmitting(true)
    try {
      const token = localStorage.getItem('token')
      await axios.put(
        `/api/comments/${commentId}`,
        { comment: editText },
        { headers: { Authorization: `Bearer ${token}` } }
      )
      setEditingId(null)
      await fetchComments()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update comment')
    } finally {
      setSubmitting(false)
    }
  }

  // Delete comment
  const deleteComment = async (commentId: number) => {
    if (!confirm('Delete this comment?')) return

    setSubmitting(true)
    try {
      const token = localStorage.getItem('token')
      await axios.delete(`/api/comments/${commentId}`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      await fetchComments()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete comment')
    } finally {
      setSubmitting(false)
    }
  }

  // Format timestamp
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    return date.toLocaleDateString()
  }

  if (!isParticipant) {
    return (
      <div className="event-comments">
        <h3>💬 Comments</h3>
        <p className="join-to-comment">Join this event to see and post comments</p>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="event-comments">
        <h3>💬 Comments</h3>
        <p>Loading comments...</p>
      </div>
    )
  }

  return (
    <div className="event-comments">
      <h3>💬 Comments ({comments.length})</h3>

      {error && <div className="error-message">{error}</div>}

      {/* Comment form */}
      <form className="comment-form" onSubmit={handleSubmit}>
        <textarea
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          placeholder="Write a comment..."
          maxLength={1000}
          rows={3}
          disabled={submitting}
        />
        <div className="form-footer">
          <span className="char-count">{newComment.length}/1000</span>
          <button type="submit" disabled={!newComment.trim() || submitting}>
            {submitting ? 'Posting...' : 'Post Comment'}
          </button>
        </div>
      </form>

      {/* Comments list */}
      <div className="comments-list">
        {comments.length === 0 ? (
          <p className="no-comments">No comments yet. Be the first to comment!</p>
        ) : (
          comments.map((comment) => (
            <div key={comment.id} className={`comment ${comment.is_own ? 'own-comment' : ''}`}>
              <div className="comment-header">
                <div className="comment-avatar">
                  {comment.user_name.charAt(0).toUpperCase()}
                </div>
                <div className="comment-meta">
                  <span className="comment-author">
                    {comment.user_name}
                    {comment.is_own && <span className="you-badge">You</span>}
                  </span>
                  <span className="comment-time">
                    {formatTime(comment.created_at)}
                    {comment.updated_at && ' (edited)'}
                  </span>
                </div>
              </div>

              <div className="comment-body">
                {editingId === comment.id ? (
                  <div className="comment-edit">
                    <textarea
                      value={editText}
                      onChange={(e) => setEditText(e.target.value)}
                      maxLength={1000}
                      rows={3}
                      disabled={submitting}
                    />
                    <div className="edit-actions">
                      <button
                        type="button"
                        onClick={cancelEdit}
                        className="cancel-btn"
                        disabled={submitting}
                      >
                        Cancel
                      </button>
                      <button
                        type="button"
                        onClick={() => saveEdit(comment.id)}
                        className="save-btn"
                        disabled={!editText.trim() || submitting}
                      >
                        {submitting ? 'Saving...' : 'Save'}
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <p className="comment-text">{comment.comment}</p>
                    {comment.is_own && (
                      <div className="comment-actions">
                        <button
                          type="button"
                          onClick={() => startEdit(comment)}
                          className="action-btn edit-btn"
                          disabled={submitting}
                        >
                          ✏️ Edit
                        </button>
                        <button
                          type="button"
                          onClick={() => deleteComment(comment.id)}
                          className="action-btn delete-btn"
                          disabled={submitting}
                        >
                          🗑️ Delete
                        </button>
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default EventComments
