import type {
  AuthResponse,
  Board,
  Reply,
  Thread,
  ThreadDetailResponse,
  ThreadListResponse,
  UploadedImage,
  User,
  UserPublic,
} from './types'

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '/api'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

let accessToken: string | null = null
export function setAccessToken(token: string | null) {
  accessToken = token
}
export function getAccessToken() {
  return accessToken
}

interface RequestOptions {
  method?: string
  body?: unknown
  headers?: Record<string, string>
}

async function rawRequest<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const isFormData = opts.body instanceof FormData
  const headers: Record<string, string> = {
    ...(!isFormData && opts.body !== undefined ? { 'Content-Type': 'application/json' } : {}),
    ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
    ...opts.headers,
  }
  let body: BodyInit | null = null
  if (isFormData) {
    body = opts.body as FormData
  } else if (opts.body !== undefined) {
    body = JSON.stringify(opts.body)
  }
  const res = await fetch(`${API_BASE}${path}`, {
    method: opts.method ?? 'GET',
    credentials: 'include',
    cache: 'no-store',
    headers,
    body,
  })
  if (res.status === 204) {
    return undefined as T
  }
  const text = await res.text()
  let data: unknown = null
  try {
    data = text ? JSON.parse(text) : null
  } catch {
    data = null
  }
  if (!res.ok) {
    const msg =
      data && typeof data === 'object' && 'error' in data
        ? String((data as { error: unknown }).error)
        : res.statusText
    throw new ApiError(res.status, msg)
  }
  return data as T
}

// Requests carry the refresh cookie automatically (credentials: include).
// On a 401 that is not itself an auth call, try refreshing the session once
// and replay the request.
async function request<T>(path: string, opts: RequestOptions = {}, allowRefresh = true): Promise<T> {
  try {
    return await rawRequest<T>(path, opts)
  } catch (err) {
    if (!allowRefresh || !(err instanceof ApiError) || err.status !== 401 || path.startsWith('/auth/')) {
      throw err
    }
    await refresh()
    return rawRequest<T>(path, opts)
  }
}

async function refresh(): Promise<User | null> {
  try {
    const res = await rawRequest<AuthResponse>('/auth/refresh', { method: 'POST' })
    setAccessToken(res.tokens.accessToken)
    return res.user
  } catch {
    setAccessToken(null)
    return null
  }
}

export const api = {
  refresh,

  // ---- auth ----
  register(username: string, email: string, password: string) {
    return rawRequest<AuthResponse>('/auth/register', {
      method: 'POST',
      body: { username, email, password },
    })
  },
  login(username: string, password: string) {
    return rawRequest<AuthResponse>('/auth/login', {
      method: 'POST',
      body: { username, password },
    })
  },
  logout() {
    return rawRequest<void>('/auth/logout', { method: 'POST' })
  },

  // ---- content ----
  boards() {
    return request<{ boards: Board[] }>('/boards')
  },
  threads(board: string, page = 1, perPage = 20) {
    return request<ThreadListResponse>(`/boards/${encodeURIComponent(board)}/threads?page=${page}&perPage=${perPage}`)
  },
  thread(id: string) {
    return request<ThreadDetailResponse>(`/threads/${encodeURIComponent(id)}`)
  },
  createThread(board: string, title: string, body: string, imageKey?: string) {
    return request<Thread>(
      `/boards/${encodeURIComponent(board)}/threads`,
      { method: 'POST', body: { title, body, imageKey: imageKey ?? '' } },
      false,
    )
  },
  reply(threadId: string, body: string, imageKey?: string, replyToId?: string) {
    return request<Reply>(
      `/threads/${encodeURIComponent(threadId)}/replies`,
      {
        method: 'POST',
        body: {
          body,
          imageKey: imageKey ?? '',
          ...(replyToId ? { replyToId } : {}),
        },
      },
      false,
    )
  },
  user(username: string) {
    return request<UserPublic>(`/users/${encodeURIComponent(username)}`)
  },
  updateProfile(data: { avatarUrl?: string; bio?: string; color?: string }) {
    return request<User>('/users/me', { method: 'PATCH', body: data })
  },

  // ---- upload ----
  upload(file: File) {
    const form = new FormData()
    form.append('file', file)
    return request<UploadedImage>('/upload', { method: 'POST', body: form }, false)
  },
}
