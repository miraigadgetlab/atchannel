import { request } from './client'
import type { PageParams, Paginated, Role, Stats, User } from '@/types/api'

export interface SetRolesResult {
  user: User
  sessions_revoked: number
  warning?: string
}

export interface SetBannedResult {
  user_id: number
  status: 'banned' | 'unbanned'
  sessions_revoked: number
}

export const adminApi = {
  listUsers(params: PageParams = {}): Promise<Paginated<User>> {
    return request('/admin/users', {
      query: { limit: params.limit, offset: params.offset },
    })
  },

  stats(): Promise<Stats> {
    return request('/admin/stats')
  },

  setRoles(userId: number, roles: Role[]): Promise<SetRolesResult> {
    return request(`/admin/users/${userId}/roles`, { method: 'PUT', body: { roles } })
  },

  setBanned(userId: number, banned: boolean): Promise<SetBannedResult> {
    return request(`/admin/users/${userId}/ban`, { method: 'PUT', body: { banned } })
  },

  deleteUser(userId: number): Promise<{ deleted: boolean }> {
    return request(`/admin/users/${userId}`, { method: 'DELETE' })
  },
}
