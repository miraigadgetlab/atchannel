import { request } from './client'
import type { TokenPair, User } from '@/types/api'

export interface RegisterInput {
  name: string
  email: string
  password: string
  /** Optional private note for the admins (owner and admins can read it). */
  note?: string
}

export interface LoginInput {
  email: string
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

  verifyEmail(token: string): Promise<{ verified: boolean }> {
    return request('/auth/verify-email', { method: 'POST', body: { token }, auth: false })
  },

  forgotPassword(email: string): Promise<{ status: string }> {
    return request('/auth/forgot-password', { method: 'POST', body: { email }, auth: false })
  },

  resetPassword(token: string, newPassword: string): Promise<{ status: string }> {
    return request('/auth/reset-password', {
      method: 'POST',
      body: { token, new_password: newPassword },
      auth: false,
    })
  },

  /** Email-free recovery: a single-use code stands in for the emailed link. */
  resetPasswordWithCode(
    email: string,
    code: string,
    newPassword: string,
  ): Promise<{ status: string }> {
    return request('/auth/reset-password-code', {
      method: 'POST',
      body: { email, code, new_password: newPassword },
      auth: false,
    })
  },
}
