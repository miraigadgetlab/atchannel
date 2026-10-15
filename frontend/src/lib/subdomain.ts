const RESERVED = new Set([
  'www',
  'api',
  'admin',
  'app',
  'static',
  'cdn',
  'assets',
  'mail',
  'status',
  'cpanel',
  'webmail',
  'ftp',
  'blog',
  'docs',
  'help',
  'support',
  'preview',
  'auth',
  'vercel',
  'opencode',
])

export function userSubdomain(): string | null {
  if (typeof window === 'undefined') return null

  const host = window.location.hostname.toLowerCase()
  const rootDomain = import.meta.env.VITE_ROOT_DOMAIN as string | undefined
  let base = (rootDomain || '').split(':')[0].toLowerCase()

  // Local development: browsers resolve *.localhost to the loopback address
  // (RFC 6761), so http://kosero.localhost:5173 behaves like a real subdomain
  // without any DNS or hosts-file setup. Fall back to that automatically.
  if (!base) {
    base = host === 'localhost' || host.endsWith('.localhost') ? 'localhost' : 'atchannel.top'
  }

  // Always treat apex, localhost, ip addresses as normal (no subdomain logic)
  if (host === base) return null
  if (!host.endsWith('.' + base)) return null

  const sub = host.slice(0, -(base.length + 1))
  if (!sub) return null
  if (sub.includes('.')) return null
  if (RESERVED.has(sub)) return null
  if (/^[0-9]+$/.test(sub)) return null

  return sub
}
