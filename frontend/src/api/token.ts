// The session tokens live in localStorage rather than memory so a page reload
// keeps the user signed in. This is a small, deliberate trade-off: the backend
// rotates refresh tokens and detects reuse, so a stolen refresh token is
// detected and burned on the next rotation, which limits the exposure window.
const ACCESS_KEY = 'atchannel.access_token'
const REFRESH_KEY = 'atchannel.refresh_token'

export const tokenStore = {
  get access(): string | null {
    return localStorage.getItem(ACCESS_KEY)
  },

  get refresh(): string | null {
    return localStorage.getItem(REFRESH_KEY)
  },

  set(accessToken: string, refreshToken: string): void {
    localStorage.setItem(ACCESS_KEY, accessToken)
    localStorage.setItem(REFRESH_KEY, refreshToken)
  },

  clear(): void {
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
  },
}
