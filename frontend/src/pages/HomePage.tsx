import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'
import type { Board } from '../lib/types'
import { Empty, ErrorMessage, Loading } from '../components/Feedback'

export default function HomePage() {
  const [boards, setBoards] = useState<Board[] | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    api
      .boards()
      .then((res) => {
        if (!cancelled) setBoards(res.boards)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'failed to load boards')
      })
    return () => { cancelled = true }
  }, [])

  return (
    <>
      <ErrorMessage>{error}</ErrorMessage>
      {!boards && !error && <Loading />}
      {boards && boards.length === 0 && <Empty>no boards yet</Empty>}
      {boards && boards.length > 0 && (
        <ul className="board-list">
          {boards.map((b) => (
            <li key={b.id}>
              <Link to={`/b/${b.slug}`} className="board-link">
                <span className="board-slug">/{b.slug}/</span>
                <span className="board-name">{b.name}</span>
              </Link>
              {b.description && <div className="board-desc">{b.description}</div>}
            </li>
          ))}
        </ul>
      )}
    </>
  )
}
