// Hand-written mirror of the Go DTOs in backend/internal/handler.
//
// Keeping these in one place means a change on the API side has exactly one
// spot to update on the client, and the compiler points at every consumer.

export type Role = 'user' | 'moderator' | 'admin'

/** Notification kinds emitted by the backend ("comment" or "like"). */
export type NotificationType = 'comment' | 'like'

/** Sort options accepted by GET /posts. */
export type PostSort = 'new' | 'old' | 'top'

/** The caller's own account record (GET /me, auth responses). */
export interface User {
  id: number
  name: string
  email: string
  roles: Role[]
  about_me: string
  avatar_url?: string
  email_verified_at?: string
  banned_at?: string
  created_at: string
  updated_at: string
}

/** The author summary embedded in posts and comments (no private fields). */
export interface PublicUser {
  id: number
  name: string
  about_me: string
  avatar_url?: string
}

/** Someone else's public profile page (GET /users/:id). */
export interface PublicProfile {
  id: number
  name: string
  about_me: string
  avatar_url?: string
  created_at: string
  post_count: number
  comment_count: number
}

export interface Channel {
  id: number
  name: string
  title: string
  description: string
  created_by?: number
  created_at: string
  updated_at: string
}

/** GET /channels/:id returns the channel plus a slice of its posts. */
export interface ChannelDetail {
  id: number
  name: string
  title: string
  description: string
  created_by?: number
  created_at: string
  post_count: number
  posts: Post[]
}

export interface Post {
  id: number
  channel_id: number
  title: string
  content: string
  user_id: number
  created_at: string
  updated_at: string
  user: PublicUser
  likes: number
  liked_by_me: boolean
  pinned: boolean
  locked: boolean
  comments?: Comment[]
}

export interface Comment {
  id: number
  post_id: number
  content: string
  user_id: number
  created_at: string
  updated_at: string
  user: PublicUser
}

export interface Notification {
  id: number
  type: NotificationType
  post_id?: number
  comment_id?: number
  read_at?: string
  created_at: string
  actor: PublicUser
}

/** Pair returned by login and refresh. `expires_at` is a Unix timestamp. */
export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_at: number
  refresh_expires_at: number
}

/** Shape of every list endpoint: {total, limit, offset, items}. */
export interface Paginated<T> {
  total: number
  limit: number
  offset: number
  items: T[]
}

/** Admin dashboard counters (GET /admin/stats). */
export interface Stats {
  users: number
  channels: number
  posts: number
  comments: number
}

export interface PageParams {
  limit?: number
  offset?: number
}
