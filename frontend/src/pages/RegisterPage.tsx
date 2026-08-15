import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../lib/auth'
import { ErrorMessage } from '../components/Feedback'

export default function RegisterPage() {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setError('')
    try {
      await register(username.trim(), email.trim(), password)
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'registration failed')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="auth-wrap">
      <h1>register</h1>
      <form className="auth-form" onSubmit={submit}>
        <input
          className="input"
          placeholder="username (3-32: a-z 0-9 _)"
          value={username}
          autoComplete="username"
          onChange={(e) => setUsername(e.target.value)}
        />
        <input
          className="input"
          placeholder="email"
          type="email"
          value={email}
          autoComplete="email"
          onChange={(e) => setEmail(e.target.value)}
        />
        <input
          className="input"
          placeholder="password (min 8 chars)"
          type="password"
          value={password}
          autoComplete="new-password"
          onChange={(e) => setPassword(e.target.value)}
        />
        <ErrorMessage>{error}</ErrorMessage>
        <button type="submit" className="btn btn-primary" disabled={busy}>
          {busy ? 'registering…' : 'register'}
        </button>
      </form>
      <p className="muted">
        have an account? <Link to="/login">login</Link>
      </p>
    </div>
  )
}
