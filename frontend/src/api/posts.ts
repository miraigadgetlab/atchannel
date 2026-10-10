import { request } from './client'
import type { PageParams, Paginated, Post, PostSort } from '@/types/api'

export interface PostListParams extends PageParams {
  q?: string
  channel_id?: number
  sort?: PostSort
}

export interface CreatePostInput {
  channel_id: number
  title: string
  content: string
}

export interface UpdatePostInput {
  title?: string
  content?: string
}

export const postsApi = {
  list(params: PostListParams = {}): Promise<Paginated<Post>> {
    return request('/posts', {
      query: {
        q: params.q,
        channel_id: params.channel_id,
        sort: params.sort,
        limit: params.limit,
        offset: params.offset,
      },
    })
  },

  get(id: number | string): Promise<Post> {
    return request(`/posts/${id}`)
  },

  create(input: CreatePostInput): Promise<Post> {
    return request('/posts', { method: 'POST', body: input })
  },

  update(id: number | string, input: UpdatePostInput): Promise<Post> {
    return request(`/posts/${id}`, { method: 'PUT', body: input })
  },

  remove(id: number | string): Promise<{ deleted: boolean }> {
    return request(`/posts/${id}`, { method: 'DELETE' })
  },

  like(id: number | string): Promise<{ liked: boolean }> {
    return request(`/posts/${id}/like`, { method: 'POST' })
  },

  unlike(id: number | string): Promise<{ liked: boolean }> {
    return request(`/posts/${id}/like`, { method: 'DELETE' })
  },

  // Moderator / admin only.
  pin(id: number | string): Promise<{ pinned: boolean }> {
    return request(`/posts/${id}/pin`, { method: 'POST' })
  },

  unpin(id: number | string): Promise<{ pinned: boolean }> {
    return request(`/posts/${id}/pin`, { method: 'DELETE' })
  },

  lock(id: number | string): Promise<{ locked: boolean }> {
    return request(`/posts/${id}/lock`, { method: 'POST' })
  },

  unlock(id: number | string): Promise<{ locked: boolean }> {
    return request(`/posts/${id}/lock`, { method: 'DELETE' })
  },
}
