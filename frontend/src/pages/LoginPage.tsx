import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../lib/auth'
import { ErrorMessage } from '../components/Feedback'

export default function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setError('')
    try {
      await login(username.trim(), password)
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'login failed')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="auth-wrap">
      <h1>login</h1>
      <form className="auth-form" onSubmit={submit}>
        <input
          className="input"
          placeholder="username"
          value={username}
          autoComplete="username"
          onChange={(e) => setUsername(e.target.value)}
        />
        <input
          className="input"
          placeholder="password"
          type="password"
          value={password}
          autoComplete="current-password"
          onChange={(e) => setPassword(e.target.value)}
        />
        <ErrorMessage>{error}</ErrorMessage>
        <button type="submit" className="btn btn-primary" disabled={busy}>
          {busy ? 'logging in…' : 'login'}
        </button>
      </form>
      <p className="muted">
        no account? <Link to="/register">register</Link>
      </p>
    </div>
  )
}
