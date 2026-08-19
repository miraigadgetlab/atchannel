export type Role = 'user' | 'mod' | 'admin'

export interface User {
  id: string
  username: string
  email: string
  avatarUrl: string
  bio: string
  role: Role
  createdAt: string
}

export interface UserPublic {
  id: string
  username: string
  avatarUrl: string
  bio: string
  role: Role
  createdAt: string
}

export interface Board {
  id: string
  slug: string
  name: string
  description: string
  createdAt: string
}

export interface Thread {
  id: string
  userId: string
  title: string
  body: string
  imageUrl: string
  isPinned: boolean
  isLocked: boolean
  bumpedAt: string
  createdAt: string
  boardSlug: string
  authorName: string
  authorRole: Role
  authorAvatar: string
  replyCount: number
  lastReplyAt?: string
  bumped: boolean
  bumpLimit: boolean
  imageArchived: boolean
  resto: string
  sticky: boolean
  closed: boolean
}

export interface Reply {
  id: string
  threadId: string
  userId: string
  body: string
  imageUrl: string
  replyToId?: string
  createdAt: string
  authorName: string
  authorRole: Role
  authorAvatar: string
  deleted: boolean
}

export interface UploadedImage {
  key: string
  url: string
  thumbKey: string
  thumbUrl: string
  width: number
  height: number
  thumbWidth: number
  thumbHeight: number
  sizeBytes: number
}

export interface TokenPair {
  accessToken: string
  refreshToken: string
  expiresIn: number
  tokenType: string
}

export interface AuthResponse {
  user: User
  tokens: TokenPair
}

export interface ThreadListResponse {
  threads: Thread[]
  total: number
}

export interface ThreadDetailResponse {
  thread: Thread
  replies: Reply[]
}
