import { AtSign } from 'lucide-react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../lib/auth'

export default function Header() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const onLogout = async () => {
    await logout()
    navigate('/')
  }

  return (
    <header className="site-header">
      <nav className="nav">
        <Link to="/" className="brand">
          <AtSign size={16} aria-hidden />
          <span>@channel</span>
        </Link>
        <div className="nav-right">
          {user ? (
            <>
              <Link to={`/users/${user.username}`} className="nav-link">
                {user.username}
              </Link>
              <button type="button" className="btn btn-ghost" onClick={onLogout}>
                logout
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="btn btn-ghost">
                login
              </Link>
              <Link to="/register" className="btn btn-primary">
                register
              </Link>
            </>
          )}
        </div>
      </nav>
    </header>
  )
}
