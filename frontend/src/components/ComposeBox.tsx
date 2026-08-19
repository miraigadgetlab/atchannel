import { useRef, useState } from 'react'
import { api } from '../lib/api'
import { ErrorMessage } from './Feedback'

interface ComposeBoxProps {
  /** 'thread' shows title+body; 'reply' shows body only */
  mode: 'thread' | 'reply'
  /** Board slug, required for thread mode */
  boardSlug?: string
  /** Thread ID, required for reply mode */
  threadId?: string
  /** Show "replying to" bar */
  replyTo?: { id: string; authorName: string } | null
  /** Called when reply-to should be cleared */
  onCancelReply?: () => void
  /** Called after successful submit with the new item */
  onPosted?: () => void
}

export default function ComposeBox({
  mode,
  boardSlug,
  threadId,
  replyTo,
  onCancelReply,
  onPosted,
}: ComposeBoxProps) {
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [posting, setPosting] = useState(false)
  const [error, setError] = useState('')

  const [imageFile, setImageFile] = useState<File | null>(null)
  const [imagePreview, setImagePreview] = useState<string | null>(null)
  const [imageKey, setImageKey] = useState<string | null>(null)
  const [uploading, setUploading] = useState(false)

  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const titleMax = 200
  const bodyMax = 4000

  const titleRemaining = titleMax - title.length
  const bodyRemaining = bodyMax - body.length

  const handleImageSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.size > 10 * 1024 * 1024) {
      setError('image must be under 10 mb')
      return
    }

    setImageFile(file)
    setImagePreview(URL.createObjectURL(file))
    setError('')
    setUploading(true)

    try {
      const uploaded = await api.upload(file)
      setImageKey(uploaded.key)

      const ta = textareaRef.current
      const marker = `[${file.name}]`
      if (ta) {
        const start = ta.selectionStart
        const end = ta.selectionEnd
        const next = body.slice(0, start) + marker + body.slice(end)
        setBody(next)
        requestAnimationFrame(() => {
          ta.selectionStart = ta.selectionEnd = start + marker.length
          ta.focus()
        })
      } else {
        setBody((prev) => prev + marker)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'upload failed')
      setImageFile(null)
      setImagePreview(null)
    } finally {
      setUploading(false)
    }
  }

  const removeImage = () => {
    if (imagePreview) URL.revokeObjectURL(imagePreview)
    if (imageFile) {
      setBody((prev) => prev.replace(`[${imageFile.name}]`, ''))
    }
    setImageFile(null)
    setImagePreview(null)
    setImageKey(null)
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (posting) return

    if (mode === 'thread') {
      if (!title.trim() || !body.trim()) return
    } else {
      if (!body.trim()) return
    }
    if (uploading) return

    setPosting(true)
    setError('')

    try {
      if (mode === 'thread' && boardSlug) {
        await api.createThread(boardSlug, title.trim(), body.trim(), imageKey ?? undefined)
        setTitle('')
        setBody('')
        removeImage()
      } else if (mode === 'reply' && threadId) {
        await api.reply(threadId, body.trim(), imageKey ?? undefined, replyTo?.id)
        setBody('')
        removeImage()
        onCancelReply?.()
      }
      onPosted?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to post')
    } finally {
      setPosting(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault()
      const form = e.currentTarget.closest('form')
      if (form) form.requestSubmit()
    }
  }

  return (
    <form className="compose-box" onSubmit={submit}>
      {replyTo && (
        <div className="compose-reply-to">
          <span className="compose-reply-to-label">
            replying to <strong>{replyTo.authorName}</strong> (No.{replyTo.id.slice(0, 8)})
          </span>
          <button type="button" className="compose-reply-to-cancel" onClick={onCancelReply}>
            &times;
          </button>
        </div>
      )}

      {mode === 'thread' && (
        <div className="compose-field">
          <input
            className="compose-input"
            placeholder="subject"
            value={title}
            maxLength={titleMax}
            onChange={(e) => setTitle(e.target.value)}
          />
          <span className={`compose-char-count${titleRemaining < 20 ? ' compose-char-count-warn' : ''}`}>
            {title.length > 0 && `${title.length}/${titleMax}`}
          </span>
        </div>
      )}

      <div className="compose-field">
        <textarea
          ref={textareaRef}
          className="compose-textarea"
          placeholder={mode === 'thread' ? 'comment' : 'write a reply...'}
          value={body}
          rows={mode === 'thread' ? 5 : 3}
          maxLength={bodyMax}
          onChange={(e) => setBody(e.target.value)}
          onKeyDown={handleKeyDown}
        />
        <span className={`compose-char-count${bodyRemaining < 200 ? ' compose-char-count-warn' : ''}`}>
          {body.length > 0 && `${body.length}/${bodyMax}`}
        </span>
      </div>

      {imagePreview && (
        <div className="compose-image-preview">
          <img src={imagePreview} alt="upload preview" />
          <button type="button" className="compose-image-remove" onClick={removeImage} title="remove image">
            &times;
          </button>
        </div>
      )}

      <div className="compose-footer">
        <div className="compose-footer-left">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            className="compose-file-input"
            onChange={handleImageSelect}
          />
          <button
            type="button"
            className="compose-attach-btn"
            onClick={() => fileInputRef.current?.click()}
            disabled={uploading || !!imageFile}
            title="attach image"
          >
            {uploading ? 'uploading...' : imageFile ? 'image attached' : '+ image'}
          </button>
          <span className="compose-shortcut-hint">ctrl+enter to submit</span>
        </div>
        <button
          type="submit"
          className="btn btn-primary compose-submit"
          disabled={posting || uploading || (mode === 'thread' ? !title.trim() || !body.trim() : !body.trim())}
        >
          {posting ? 'posting...' : mode === 'thread' ? 'post thread' : 'reply'}
        </button>
      </div>

      <ErrorMessage>{error}</ErrorMessage>
    </form>
  )
}
