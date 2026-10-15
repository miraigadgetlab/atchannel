import { request } from './client'
import type { TokenPair, User } from '@/types/api'

export interface RegisterInput {
  name: string
  password: string
  /** Optional private note for the admins (owner and admins can read it). */
  note?: string
}

export interface LoginInput {
  name: string
  password: string
}

export const authApi = {
  register(input: RegisterInput): Promise<User> {
    return request('/auth/register', { method: 'POST', body: input, auth: false })
  },

  login(input: LoginInput): Promise<TokenPair> {
    return request('/auth/login', { method: 'POST', body: input, auth: false })
  },

  logout(refreshToken: string): Promise<{ revoked: boolean }> {
    return request('/auth/logout', {
      method: 'POST',
      body: { refresh_token: refreshToken },
      auth: false,
    })
  },

  logoutAll(): Promise<{ revoked_count: number }> {
    return request('/auth/logout-all', { method: 'POST' })
  },

  /** Code-based recovery: a single-use code, generated in settings, stands in
   * for an emailed reset link. */
  resetPasswordWithCode(
    name: string,
    code: string,
    newPassword: string,
  ): Promise<{ status: string }> {
    return request('/auth/reset-password-code', {
      method: 'POST',
      body: { name, code, new_password: newPassword },
      auth: false,
    })
  },
}
