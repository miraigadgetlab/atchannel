import { API_BASE_URL } from '@/config'
import { tokenStore } from './token'
import type { TokenPair } from '@/types/api'

/**
 * Error thrown for any non-2xx response. `message` is the backend's
 * `{ "message": ... }` field when present, so callers can show it verbatim.
 */
export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  query?: Record<string, string | number | undefined>
  /** Attach the access token and enable the refresh-on-401 flow. Default true. */
  auth?: boolean
  headers?: Record<string, string>
}

export const AUTH_UNAUTHORIZED_EVENT = 'auth:unauthorized'

function buildUrl(path: string, query?: RequestOptions['query']): string {
  const url = new URL(API_BASE_URL.replace(/\/$/, '') + path, window.location.origin)

  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined && value !== '') {
        url.searchParams.set(key, String(value))
      }
    }
  }

  return url.toString()
}

async function parseError(res: Response): Promise<ApiError> {
  let message = res.statusText || `Request failed (${res.status})`

  try {
    const data: unknown = await res.json()
    if (data && typeof data === 'object' && 'message' in data) {
      const detail = (data as { message?: unknown }).message
      if (typeof detail === 'string' && detail) {
        message = detail
      }
    }
  } catch {
    // Non-JSON body (proxy error, empty response): keep the generic message.
  }

  return new ApiError(res.status, message)
}

// A single in-flight refresh shared by every request that hits a 401 at once,
// so a burst of parallel calls rotates the token only once.
let refreshPromise: Promise<boolean> | null = null

async function refreshTokens(): Promise<boolean> {
  const refreshToken = tokenStore.refresh
  if (!refreshToken) {
    return false
  }

  try {
    const res = await fetch(buildUrl('/auth/refresh'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!res.ok) {
      return false
    }

    const pair = (await res.json()) as TokenPair
    tokenStore.set(pair.access_token, pair.refresh_token)
    return true
  } catch {
    return false
  }
}

function invalidateSession(): void {
  tokenStore.clear()
  window.dispatchEvent(new CustomEvent(AUTH_UNAUTHORIZED_EVENT))
}

/**
 * Thin fetch wrapper: JSON in/out, bearer auth, and one transparent token
 * refresh before giving up on a 401.
 */
export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, query, auth = true, headers } = options
  const url = buildUrl(path, query)
  const isFormData = body instanceof FormData

  const send = (): Promise<Response> => {
    const finalHeaders: Record<string, string> = { Accept: 'application/json', ...headers }

    if (body !== undefined && !isFormData) {
      finalHeaders['Content-Type'] = 'application/json'
    }
    if (auth) {
      const token = tokenStore.access
      if (token) {
        finalHeaders.Authorization = `Bearer ${token}`
      }
    }

    return fetch(url, {
      method,
      headers: finalHeaders,
      body: body === undefined ? undefined : isFormData ? body : JSON.stringify(body),
    })
  }

  let res = await send()

  if (res.status === 401 && auth && tokenStore.refresh) {
    refreshPromise ??= refreshTokens().finally(() => {
      refreshPromise = null
    })

    const refreshed = await refreshPromise
    if (refreshed) {
      res = await send()
    }
  }

  if (!res.ok) {
    const error = await parseError(res)

    // A 401 on an authenticated request means the session is gone (refresh
    // failed or was absent). Clear it once so the app can react.
    if (res.status === 401 && auth) {
      invalidateSession()
    }

    throw error
  }

  if (res.status === 204) {
    return undefined as T
  }

  const text = await res.text()
  return text ? (JSON.parse(text) as T) : (undefined as T)
}
