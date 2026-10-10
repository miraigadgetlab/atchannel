import { UPLOADS_BASE_URL } from '@/config'

/**
 * Avatars come back from the API as root-relative paths ("/uploads/x.png").
 * In development the Vite proxy serves them from the same origin, so no prefix
 * is needed; behind a split deployment VITE_UPLOADS_BASE_URL can point them at
 * the backend origin instead.
 */
export function mediaUrl(path?: string): string | undefined {
  if (!path) {
    return undefined
  }
  if (/^https?:\/\//i.test(path)) {
    return path
  }
  return UPLOADS_BASE_URL.replace(/\/$/, '') + path
}
