import { useEffect, useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../lib/api'
import type { UserPublic } from '../lib/types'
import { ErrorMessage, Loading } from '../components/Feedback'
import { formatDate } from '../lib/format'
import { useAuth } from '../lib/auth'

const COLORS = [
  '#7C3AED', '#3B82F6', '#10B981',
  '#8B5CF6', '#06B6D4', '#059669',
  '#6366F1', '#0EA5E9', '#14B8A6',
]

export default function UserPage() {
  const { username } = useParams<{ username: string }>()
  const { user: currentUser, updateUser } = useAuth()
  const [profile, setProfile] = useState<UserPublic | null>(null)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [selectedColor, setSelectedColor] = useState('')

  const isOwn = currentUser?.username === username

  useEffect(() => {
    let cancelled = false
    setProfile(null)
    api
      .user(username ?? '')
      .then((u) => {
        if (!cancelled) {
          setProfile(u)
          setSelectedColor(u.color || COLORS[0])
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'user not found')
      })
    return () => { cancelled = true }
  }, [username])

  const saveColor = useCallback(async () => {
    if (!profile) return
    setSaving(true)
    setSaved(false)
    try {
      const updated = await api.updateProfile({ color: selectedColor })
      setProfile({ ...profile, color: updated.color })
      updateUser({ color: updated.color })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch {
      setError('failed to update color')
    } finally {
      setSaving(false)
    }
  }, [profile, selectedColor, updateUser])

  return (
    <>
      <ErrorMessage>{error}</ErrorMessage>
      {!profile && !error && <Loading />}
      {profile && (
        <div className="profile">
          {profile.avatarUrl && <img src={profile.avatarUrl} alt="" className="avatar-large" />}
          <div className="post-header profile-header">
            <span className="post-name">{profile.username}</span>
            <span className="post-seq">role: {profile.role}</span>
            <span className="post-timestamp">joined {formatDate(profile.createdAt)}</span>
          </div>
          {profile.bio && <p className="post-body">{profile.bio}</p>}

          {isOwn && (
            <div className="color-picker-section">
              <div className="color-picker-label">post color</div>
              <div className="color-picker-row">
                {COLORS.map((c) => (
                  <button
                    key={c}
                    className={`color-swatch${selectedColor === c ? ' color-swatch-active' : ''}`}
                    style={{ background: c }}
                    onClick={() => setSelectedColor(c)}
                    type="button"
                  />
                ))}
              </div>
              {selectedColor !== (profile.color || COLORS[0]) && (
                <button
                  className="btn color-save-btn"
                  onClick={saveColor}
                  disabled={saving}
                  type="button"
                >
                  {saving ? 'saving...' : 'save'}
                </button>
              )}
              {saved && <span className="color-saved-msg">saved!</span>}
            </div>
          )}
        </div>
      )}
    </>
  )
}
