import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'
import type { Thread } from '../lib/types'
import { Empty, ErrorMessage, Loading } from '../components/Feedback'
import ComposeBox from '../components/ComposeBox'
import { useAuth } from '../lib/auth'
import { formatDate } from '../lib/format'

export default function BoardPage() {
  const { board } = useParams<{ board: string }>()
  const slug = board ?? ''
  const { user } = useAuth()
  const [threads, setThreads] = useState<Thread[] | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    setThreads(null)
    api
      .threads(slug)
      .then((res) => {
        if (cancelled) return
        setThreads(res.threads)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'failed to load threads')
      })
    return () => { cancelled = true }
  }, [slug])

  const load = async () => {
    const res = await api.threads(slug)
    setThreads(res.threads)
  }

  return (
    <>
      <h1>/{slug}/</h1>
      <ErrorMessage>{error}</ErrorMessage>

      {user && (
        <ComposeBox
          mode="thread"
          boardSlug={slug}
          onPosted={load}
        />
      )}

      {!threads && !error && <Loading />}
      {threads && threads.length === 0 && <Empty>no threads here yet</Empty>}
      {threads && threads.length > 0 && (
        <ul className="thread-list">
          {threads.map((t, i) => (
            <li key={t.id} className={`thread-item${t.isPinned ? ' thread-item-pinned' : ''}${t.isLocked ? ' thread-item-locked' : ''}`}>
              <span className="thread-num">{i + 1}.</span>
              {t.isPinned && <span className="thread-tag thread-tag-pinned">[PINNED]</span>}
              {t.isLocked && <span className="thread-tag">[LOCKED]</span>}
              <Link to={`/t/${t.id}`} className="thread-title-link">
                {t.title}
              </Link>
              <span className="thread-replycount">({t.replyCount})</span>
              {t.body && (
                <span className="thread-preview">
                  {t.body.length > 100 ? t.body.slice(0, 100) + '...' : t.body}
                </span>
              )}
              <span className="thread-bumped">{formatDate(t.bumpedAt)}</span>
              {t.imageUrl && (
                <img src={t.imageUrl} alt="" className="thread-thumb" loading="lazy" />
              )}
            </li>
          ))}
        </ul>
      )}
    </>
  )
}
