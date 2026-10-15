import type { RouteLocationRaw } from 'vue-router'

import { getBaseUrl } from '@/lib/seo'

// Single-segment paths that belong to the app or to the host rather than to an
// account. A user named "login" would be shadowed by the /login route, and
// /favicon.svg never even reaches the SPA because the host serves the file
// first. Accounts whose name lands here are linked through /users/:id instead
// of being silently unreachable.
const RESERVED = new Set([
  // App pages (see src/router/index.ts).
  'admin',
  'channels',
  'channelers',
  'forgot-password',
  'login',
  'me',
  'notifications',
  'posts',
  'register',
  'users',
  // Reserved for the API/upload rewrites and the build output.
  'api',
  'assets',
  'fonts',
  'uploads',
  // Static files in public/, served before the SPA fallback runs.
  'favicon.ico',
  'favicon.svg',
  'index.html',
  'manifest.webmanifest',
  'og-image.png',
  'robots.txt',
  'site.webmanifest',
  'sitemap.xml',
])

/** True when `name` cannot be used as a root-level profile path. */
export function isReservedProfileName(name: string): boolean {
  return RESERVED.has(name.trim().toLowerCase())
}

/** The account's name when it can live at /<name>, otherwise null. */
function profileName(user: { name: string }): string | null {
  const name = user.name.trim()
  return name && !isReservedProfileName(name) ? name : null
}

/** Router target for someone's public profile: /emirhan, or /users/4. */
export function profileRoute(user: { id: number; name: string }): RouteLocationRaw {
  const name = profileName(user)
  return name
    ? { name: 'profile-by-name', params: { name } }
    : { name: 'profile', params: { id: user.id } }
}

/** Absolute, shareable URL for a profile; used as the canonical/OG url. */
export function profileUrl(user: { id: number; name: string }): string {
  const name = profileName(user)
  return name ? `${getBaseUrl()}/${encodeURIComponent(name)}` : `${getBaseUrl()}/users/${user.id}`
}
