import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../lib/api'
import type { UserPublic } from '../lib/types'
import { ErrorMessage, Loading } from '../components/Feedback'
import { formatDate } from '../lib/format'

export default function UserPage() {
  const { username } = useParams<{ username: string }>()
  const [user, setUser] = useState<UserPublic | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    setUser(null)
    api
      .user(username ?? '')
      .then((u) => {
        if (!cancelled) setUser(u)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'user not found')
      })
    return () => { cancelled = true }
  }, [username])

  return (
    <>
      <ErrorMessage>{error}</ErrorMessage>
      {!user && !error && <Loading />}
      {user && (
        <div className="profile">
          {user.avatarUrl && <img src={user.avatarUrl} alt="" className="avatar-large" />}
          <div className="post-header profile-header">
            <span className="post-name">{user.username}</span>
            <span className="post-seq">role: {user.role}</span>
            <span className="post-timestamp">joined {formatDate(user.createdAt)}</span>
          </div>
          {user.bio && <p className="post-body">{user.bio}</p>}
        </div>
      )}
    </>
  )
}
