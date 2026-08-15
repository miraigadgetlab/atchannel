import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'
import type { Reply, Thread } from '../lib/types'
import { Empty, ErrorMessage, Loading } from '../components/Feedback'
import { useAuth } from '../lib/auth'
import { formatDate } from '../lib/format'

export default function ThreadPage() {
  const { id } = useParams<{ id: string }>()
  const threadId = id ?? ''
  const { user } = useAuth()
  const [thread, setThread] = useState<Thread | null>(null)
  const [replies, setReplies] = useState<Reply[] | null>(null)
  const [error, setError] = useState('')
  const [body, setBody] = useState('')
  const [replyTo, setReplyTo] = useState<string | null>(null)
  const [posting, setPosting] = useState(false)
  const [postError, setPostError] = useState('')

  useEffect(() => {
    let cancelled = false
    setThread(null)
    setReplies(null)
    api
      .thread(threadId)
      .then((res) => {
        if (cancelled) return
        setThread(res.thread)
        setReplies(res.replies)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'failed to load thread')
      })
    return () => {
      cancelled = true
    }
  }, [threadId])

  const load = async () => {
    const res = await api.thread(threadId)
    setThread(res.thread)
    setReplies(res.replies)
  }

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!body.trim()) return
    setPosting(true)
    setPostError('')
    try {
      await api.reply(threadId, body.trim(), undefined, replyTo ?? undefined)
      setBody('')
      setReplyTo(null)
      // Reload so the new reply is enriched with author aggregates.
      await load()
    } catch (err) {
      setPostError(err instanceof Error ? err.message : 'failed to post reply')
    } finally {
      setPosting(false)
    }
  }

  return (
    <>
      <Link to={thread ? `/b/${thread.boardSlug}` : '/'} className="muted">
        ← back to board
      </Link>
      <ErrorMessage>{error}</ErrorMessage>

      {!thread && !error && <Loading />}

      {thread && (
        <article className="thread-post">
          <h2>{thread.title}</h2>
          {thread.imageUrl && <img src={thread.imageUrl} alt="" className="post-image" />}
          <p className="post-body">{thread.body}</p>
          <div className="muted post-meta">
            by <Link to={`/users/${thread.authorName}`}>{thread.authorName}</Link> ·{' '}
            {formatDate(thread.createdAt)} {thread.closed && '· locked'}
          </div>
        </article>
      )}

      {user && !error && thread && !thread.closed && (
        <form className="post-form" onSubmit={submit}>
          {replyTo && (
            <div className="muted">
              replying to <button type="button" className="btn-link" onClick={() => setReplyTo(null)}>#{replyTo} (cancel)</button>
            </div>
          )}
          <textarea
            className="input"
            placeholder="reply"
            value={body}
            rows={3}
            maxLength={4000}
            onChange={(e) => setBody(e.target.value)}
          />
          <ErrorMessage>{postError}</ErrorMessage>
          <div className="form-actions">
            <button type="submit" className="btn btn-primary" disabled={posting}>
              {posting ? 'posting…' : 'post reply'}
            </button>
          </div>
        </form>
      )}

      {!replies && !error && <Loading />}
      {replies && replies.length === 0 && <Empty>no replies yet</Empty>}
      {replies && replies.length > 0 && (
        <ul className="reply-list">
          {replies.map((r) => (
            <li key={r.id} className="reply-item">
              <div className="muted post-meta">
                <button
                  type="button"
                  className="btn-link"
                  onClick={() => user && setReplyTo(r.id)}
                  title={user ? 'reply to this post' : 'log in to reply'}
                >
                  #{r.id.slice(0, 8)}
                </button>{' '}
                by <Link to={`/users/${r.authorName}`}>{r.authorName}</Link> ·{' '}
                {formatDate(r.createdAt)}
                {r.replyToId && <> · in reply to #{r.replyToId.slice(0, 8)}</>}
              </div>
              {r.imageUrl && <img src={r.imageUrl} alt="" className="post-image" />}
              <p className="post-body">{r.body}</p>
            </li>
          ))}
        </ul>
      )}
    </>
  )
}
