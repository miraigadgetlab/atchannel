import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { authApi } from '@/api/auth'
import { usersApi } from '@/api/users'
import type { ChangePasswordResult, UpdateProfileInput } from '@/api/users'
import { AUTH_UNAUTHORIZED_EVENT } from '@/api/client'
import { tokenStore } from '@/api/token'
import type { Role, TokenPair, User } from '@/types/api'

/**
 * The single source of truth for "who is signed in".
 *
 * Tokens are persisted (see api/token.ts) but the authoritative user record is
 * always re-fetched from GET /me: roles and ban state can change server-side,
 * and a token carries a snapshot of the roles it was minted with.
 */
export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const initialized = ref(false)

  const isAuthenticated = computed(() => user.value !== null)
  const roles = computed<Role[]>(() => user.value?.roles ?? [])
  const isAdmin = computed(() => roles.value.includes('admin'))
  const canModerate = computed(
    () => roles.value.includes('admin') || roles.value.includes('moderator'),
  )
  const isEmailVerified = computed(() => user.value?.email_verified_at != null)

  function clearSession(): void {
    tokenStore.clear()
    user.value = null
  }

  function applyTokens(pair: TokenPair): void {
    tokenStore.set(pair.access_token, pair.refresh_token)
  }

  async function fetchMe(): Promise<User> {
    user.value = await usersApi.me()
    return user.value
  }

  /**
   * Restores a session on app start exactly once. A stale token that the
   * backend rejects is cleared rather than retried, so the app never gets
   * stuck half-authenticated.
   */
  async function initialize(): Promise<void> {
    if (initialized.value) {
      return
    }

    if (tokenStore.access || tokenStore.refresh) {
      try {
        await fetchMe()
      } catch {
        clearSession()
      }
    }

    initialized.value = true
  }

  async function login(email: string, password: string): Promise<void> {
    const pair = await authApi.login({ email, password })
    applyTokens(pair)
    await fetchMe()
  }

  async function register(name: string, email: string, password: string): Promise<void> {
    await authApi.register({ name, email, password })
    // Registration returns the account but no tokens, so sign in right away.
    await login(email, password)
  }

  async function logout(): Promise<void> {
    const refreshToken = tokenStore.refresh
    if (refreshToken) {
      try {
        await authApi.logout(refreshToken)
      } catch {
        // Revoking server-side is best effort; the local session goes anyway.
      }
    }
    clearSession()
  }

  async function logoutAll(): Promise<void> {
    try {
      await authApi.logoutAll()
    } catch {
      // Same as logout: local state is cleared regardless.
    }
    clearSession()
  }

  async function updateProfile(input: UpdateProfileInput): Promise<User> {
    user.value = await usersApi.updateMe(input)
    return user.value
  }

  function changePassword(
    currentPassword: string,
    newPassword: string,
  ): Promise<ChangePasswordResult> {
    return usersApi.changePassword(currentPassword, newPassword)
  }

  async function uploadAvatar(file: File): Promise<User> {
    user.value = await usersApi.uploadAvatar(file)
    return user.value
  }

  function resendVerification(): Promise<{ sent: boolean }> {
    return usersApi.resendVerification()
  }

  // The HTTP client hides token refreshes, so a failed session surfaces here as
  // an event. Dropping the user object flips every guard to "signed out".
  if (typeof window !== 'undefined') {
    window.addEventListener(AUTH_UNAUTHORIZED_EVENT, () => {
      user.value = null
    })
  }

  return {
    user,
    initialized,
    isAuthenticated,
    roles,
    isAdmin,
    canModerate,
    isEmailVerified,
    initialize,
    fetchMe,
    login,
    register,
    logout,
    logoutAll,
    updateProfile,
    changePassword,
    uploadAvatar,
    resendVerification,
    clearSession,
  }
})
