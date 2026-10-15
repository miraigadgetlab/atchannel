import { request } from './client'
import type { Channeler, PageParams, Paginated, PublicProfile, User } from '@/types/api'

/** Only the provided fields are sent; the backend treats omitted as "leave". */
export interface UpdateProfileInput {
  name?: string
  about_me?: string
  avatar_url?: string
  /** Private note for the admins; empty string clears it. */
  note?: string
  /** Required by the backend to change the name (the login identity). */
  current_password?: string
}

export interface ChangePasswordResult {
  changed: boolean
  sessions_revoked: number
  warning?: string
}

export const usersApi = {
  me(): Promise<User> {
    return request('/me')
  },

  updateMe(input: UpdateProfileInput): Promise<User> {
    return request('/me', { method: 'PUT', body: input })
  },

  changePassword(currentPassword: string, newPassword: string): Promise<ChangePasswordResult> {
    return request('/me/password', {
      method: 'PUT',
      body: { current_password: currentPassword, new_password: newPassword },
    })
  },

  /** How many single-use recovery codes the caller has left. */
  recoveryCodeStatus(): Promise<{ remaining: number }> {
    return request('/me/recovery-codes')
  },

  /**
   * Generates a new batch of recovery codes and returns the plaintext exactly
   * once. Retires every previous code.
   */
  generateRecoveryCodes(): Promise<{ codes: string[] }> {
    return request('/me/recovery-codes', { method: 'POST' })
  },

  uploadAvatar(file: File): Promise<User> {
    const form = new FormData()
    form.append('avatar', file)
    return request('/me/avatar', { method: 'POST', body: form })
  },

  getProfile(id: number | string): Promise<PublicProfile> {
    return request(`/users/${id}`)
  },

  getProfileByName(name: string): Promise<PublicProfile> {
    return request(`/users/by-name/${encodeURIComponent(name)}`)
  },

  /** Public leaderboard: accounts ranked by post count, most active first. */
  listChannelers(params: PageParams = {}): Promise<Paginated<Channeler>> {
    return request('/users', {
      query: { limit: params.limit, offset: params.offset },
    })
  },
}
