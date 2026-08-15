import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'
import type { Thread } from '../lib/types'
import { Empty, ErrorMessage, Loading } from '../components/Feedback'
import { useAuth } from '../lib/auth'
import { formatDate } from '../lib/format'

export default function BoardPage() {
  const { board } = useParams<{ board: string }>()
  const slug = board ?? ''
  const { user } = useAuth()
  const [threads, setThreads] = useState<Thread[] | null>(null)
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [posting, setPosting] = useState(false)
  const [postError, setPostError] = useState('')

  useEffect(() => {
    let cancelled = false
    setThreads(null)
    api
      .threads(slug)
      .then((res) => {
        if (cancelled) return
        setThreads(res.threads)
        setTotal(res.total)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'failed to load threads')
      })
    return () => {
      cancelled = true
    }
  }, [slug])

  const load = async () => {
    const res = await api.threads(slug)
    setThreads(res.threads)
    setTotal(res.total)
  }

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim() || !body.trim()) return
    setPosting(true)
    setPostError('')
    try {
      await api.createThread(slug, title.trim(), body.trim())
      setTitle('')
      setBody('')
      // Reload so the new thread is enriched with author/board aggregates.
      await load()
    } catch (err) {
      setPostError(err instanceof Error ? err.message : 'failed to post thread')
    } finally {
      setPosting(false)
    }
  }

  return (
    <>
      <h1>
        /{slug}/ threads <span className="muted">({total})</span>
      </h1>
      <ErrorMessage>{error}</ErrorMessage>

      {user && (
        <form className="post-form" onSubmit={submit}>
          <input
            className="input"
            placeholder="title"
            value={title}
            maxLength={200}
            onChange={(e) => setTitle(e.target.value)}
          />
          <textarea
            className="input"
            placeholder="body"
            value={body}
            rows={4}
            maxLength={4000}
            onChange={(e) => setBody(e.target.value)}
          />
          <ErrorMessage>{postError}</ErrorMessage>
          <div className="form-actions">
            <button type="submit" className="btn btn-primary" disabled={posting}>
              {posting ? 'posting…' : 'start thread'}
            </button>
          </div>
        </form>
      )}

      {!threads && !error && <Loading />}
      {threads && threads.length === 0 && <Empty>no threads here yet</Empty>}
      {threads && threads.length > 0 && (
        <ul className="thread-list">
          {threads.map((t) => (
            <li key={t.id} className="thread-item">
              <Link to={`/t/${t.id}`} className="thread-title">
                {t.title}
              </Link>
              <div className="muted thread-meta">
                by {t.authorName} · {t.replyCount} replies · {formatDate(t.bumpedAt)}
              </div>
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
