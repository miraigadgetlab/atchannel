import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../lib/auth'
import { api } from '../lib/api'
import type { Board } from '../lib/types'

export default function Header() {
  const { user, logout } = useAuth()
  const [boards, setBoards] = useState<Board[]>([])

  useEffect(() => {
    let cancelled = false
    api.boards().then((res) => {
      if (!cancelled) setBoards(res.boards)
    }).catch(() => {})
    return () => { cancelled = true }
  }, [])

  return (
    <header className="site-header">
      <nav className="nav">
        <Link to="/" className="brand">@channel</Link>
        <div className="board-quicklinks">
          {boards.map((b, i) => (
            <span key={b.id}>
              <Link to={`/b/${b.slug}`}>[{b.slug}]</Link>
              {i < boards.length - 1 && ' '}
            </span>
          ))}
        </div>
        <div className="nav-right">
          {user ? (
            <>
              <Link to={`/users/${user.username}`} className="nav-link">
                {user.username}
              </Link>
              <button type="button" className="btn btn-ghost" onClick={() => { logout() }}>
                logout
              </button>
            </>
          ) : (
            <>
              <Link to="/login">login</Link>
              <Link to="/register">register</Link>
            </>
          )}
        </div>
      </nav>
    </header>
  )
}
