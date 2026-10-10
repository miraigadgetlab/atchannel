import { request } from './client'
import type { Notification, PageParams, Paginated } from '@/types/api'

export interface UnreadCount {
  unread: number
}

export const notificationsApi = {
  list(params: PageParams = {}): Promise<Paginated<Notification>> {
    return request('/notifications', {
      query: { limit: params.limit, offset: params.offset },
    })
  },

  unreadCount(): Promise<UnreadCount> {
    return request('/notifications/unread-count')
  },

  markAllRead(): Promise<{ updated: number }> {
    return request('/notifications/read', { method: 'POST' })
  },
}
