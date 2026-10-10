import { request } from './client'
import type { Comment, PageParams, Paginated } from '@/types/api'

export const commentsApi = {
  list(postId: number | string, params: PageParams = {}): Promise<Paginated<Comment>> {
    return request(`/posts/${postId}/comments`, {
      query: { limit: params.limit, offset: params.offset },
    })
  },

  create(postId: number | string, content: string): Promise<Comment> {
    return request(`/posts/${postId}/comments`, { method: 'POST', body: { content } })
  },

  remove(commentId: number | string): Promise<{ deleted: boolean }> {
    return request(`/comments/${commentId}`, { method: 'DELETE' })
  },
}
