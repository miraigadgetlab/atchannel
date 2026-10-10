import { ApiError } from '@/api/client'

/** Turns anything thrown into a string safe to show next to a form. */
export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    return err.message
  }
  if (err instanceof Error) {
    return err.message
  }
  return 'Something went wrong. Please try again.'
}

/** 404s are common enough (deleted posts/users) to deserve a helper. */
export function isNotFound(err: unknown): boolean {
  return err instanceof ApiError && err.status === 404
}
