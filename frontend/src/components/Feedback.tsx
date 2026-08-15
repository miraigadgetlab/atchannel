import type { ReactNode } from 'react'

export function ErrorMessage({ children }: { children: ReactNode }) {
  if (!children) return null
  return <div className="error">{children}</div>
}

export function Loading() {
  return <div className="muted">loading…</div>
}

export function Empty({ children }: { children: ReactNode }) {
  return <div className="muted">{children}</div>
}
