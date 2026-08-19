import React, { useEffect, useState, useRef, useCallback, useMemo } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'
import type { Reply, Thread } from '../lib/types'
import { Empty, ErrorMessage, Loading } from '../components/Feedback'
import { useAuth } from '../lib/auth'
import { formatDate } from '../lib/format'
import defaultAvatar from '../assets/avatar.png'

function renderBody(text: string, allPosts: { id: string; body: string; authorName: string }[]): React.ReactNode[] {
  const parts: React.ReactNode[] = []
  const regex = />>(\w{8,})/g
  let lastIndex = 0
  let match: RegExpExecArray | null

  while ((match = regex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push(text.slice(lastIndex, match.index))
    }
    const refId = match[1]
    const referenced = allPosts.find((p) => p.id.startsWith(refId))
    if (referenced) {
      parts.push(
        <ReplyRef key={match.index} refId={refId} fullId={referenced.id} body={referenced.body} authorName={referenced.authorName} />
      )
    } else {
      parts.push(<span key={match.index} className="reply-ref">&gt;&gt;{refId}</span>)
    }
    lastIndex = match.index + match[0].length
  }
  if (lastIndex < text.length) {
    parts.push(text.slice(lastIndex))
  }
  return parts
}

function ReplyRef({ refId, body, authorName }: { refId: string; fullId: string; body: string; authorName: string }) {
  const [show, setShow] = useState(false)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    return () => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current)
    }
  }, [])

  const enter = useCallback(() => {
    if (timeoutRef.current) clearTimeout(timeoutRef.current)
    setShow(true)
  }, [])

  const leave = useCallback(() => {
    timeoutRef.current = setTimeout(() => setShow(false), 150)
  }, [])

  return (
    <span
      className="reply-ref"
      onMouseEnter={enter}
      onMouseLeave={leave}
    >
      &gt;&gt;{refId}
      {show && (
        <span className="reply-ref-preview" onMouseEnter={enter} onMouseLeave={leave}>
          <span className="preview-name">{authorName}</span>
          <span className="preview-body">{body.length > 200 ? body.slice(0, 200) + '...' : body}</span>
        </span>
      )}
    </span>
  )
}

function PostBox({
  id,
  authorName,
  authorRole,
  authorAvatar,
  body,
  createdAt,
  imageUrl,
  replyToId,
  allPosts,
  isOp,
  onReply,
  user,
}: {
  id: string
  authorName: string
  authorRole: string
  authorAvatar: string
  body: string
  createdAt: string
  imageUrl: string
  replyToId?: string
  allPosts: { id: string; body: string; authorName: string }[]
  isOp?: boolean
  onReply?: () => void
  user?: { username: string } | null
}) {
  const repliedTo = replyToId ? allPosts.find((p) => p.id === replyToId) : null

  return (
    <div className={`post ${isOp ? 'post-op' : 'post-reply'}`}>
      <img
        src={authorAvatar || defaultAvatar}
        alt=""
        className="post-avatar"
        onError={(e) => { e.currentTarget.src = defaultAvatar }}
      />
      <div className="post-content">
        <div className="post-header-line">
          <Link to={`/users/${authorName}`} className="post-name">{authorName}</Link>
          {(authorRole === 'admin' || authorRole === 'mod') && (
            <span className="post-role">[{authorRole}]</span>
          )}
          {onReply && (
            <button
              type="button"
              className="btn-link"
              onClick={onReply}
              title={user ? 'reply to this post' : 'log in to reply'}
            >
              [reply]
            </button>
          )}
          <a href={`#${id.slice(0, 8)}`} className="post-id">No.{id.slice(0, 8)}</a>
        </div>
        {repliedTo && (
          <div className="reply-in-reply">
            {'\u21A9'} replying to{' '}
            <a href={`#${replyToId!.slice(0, 8)}`}>
              {repliedTo.authorName} (No.{replyToId!.slice(0, 8)})
            </a>
          </div>
        )}
        {imageUrl && <img src={imageUrl} alt="" className="post-image" />}
        <div className="post-message">
          {renderBody(body, allPosts)}
        </div>
        <div className="post-date">{formatDate(createdAt)}</div>
      </div>
    </div>
  )
}

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
    return () => { cancelled = true }
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
      await load()
    } catch (err) {
      setPostError(err instanceof Error ? err.message : 'failed to post reply')
    } finally {
      setPosting(false)
    }
  }

  const allPosts = useMemo(() => [
    ...(thread ? [{ id: thread.id, body: thread.body, authorName: thread.authorName }] : []),
    ...(replies?.map((r) => ({ id: r.id, body: r.body, authorName: r.authorName })) ?? []),
  ], [thread, replies])

  return (
    <>
      <Link to={thread ? `/b/${thread.boardSlug}` : '/'} className="back-link">
        &larr; back to board
      </Link>
      <ErrorMessage>{error}</ErrorMessage>

      {!thread && !error && <Loading />}

      {thread && (
        <PostBox
          id={thread.id}
          authorName={thread.authorName}
          authorRole={thread.authorRole}
          authorAvatar={thread.authorAvatar}
          body={thread.body}
          createdAt={thread.createdAt}
          imageUrl={thread.imageUrl}
          allPosts={allPosts}
          isOp
        />
      )}

      {user && !error && thread && !thread.closed && (
        <form className="post-form" onSubmit={submit}>
          {replyTo && (
            <div className="reply-to-bar">
              replying to #{replyTo.slice(0, 8)}{' '}
              <button type="button" className="btn-link" onClick={() => setReplyTo(null)}>(cancel)</button>
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
              {posting ? 'posting...' : 'post reply'}
            </button>
          </div>
        </form>
      )}

      {!replies && !error && <Loading />}
      {replies && replies.length === 0 && <Empty>no replies yet</Empty>}
      {replies && replies.length > 0 && (
        <div className="reply-list">
          {replies.map((r) => (
            <PostBox
              key={r.id}
              id={r.id}
              authorName={r.authorName}
              authorRole={r.authorRole}
              authorAvatar={r.authorAvatar}
              body={r.body}
              createdAt={r.createdAt}
              imageUrl={r.imageUrl}
              replyToId={r.replyToId}
              allPosts={allPosts}
              onReply={() => user && setReplyTo(r.id)}
              user={user}
            />
          ))}
        </div>
      )}
    </>
  )
}
