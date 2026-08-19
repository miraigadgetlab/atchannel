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
  const [, setTotal] = useState(0)
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
    return () => { cancelled = true }
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
      await load()
    } catch (err) {
      setPostError(err instanceof Error ? err.message : 'failed to post thread')
    } finally {
      setPosting(false)
    }
  }

  return (
    <>
      <h1>/{slug}/</h1>
      <ErrorMessage>{error}</ErrorMessage>

      {user && (
        <form className="post-form" onSubmit={submit}>
          <div className="form-row">
            <label>title</label>
            <input
              className="input"
              placeholder="thread title"
              value={title}
              maxLength={200}
              onChange={(e) => setTitle(e.target.value)}
            />
          </div>
          <div className="form-row">
            <label>body</label>
            <textarea
              className="input"
              placeholder="thread body"
              value={body}
              rows={4}
              maxLength={4000}
              onChange={(e) => setBody(e.target.value)}
            />
          </div>
          <ErrorMessage>{postError}</ErrorMessage>
          <div className="form-actions">
            <button type="submit" className="btn btn-primary" disabled={posting}>
              {posting ? 'posting...' : 'post'}
            </button>
          </div>
        </form>
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
